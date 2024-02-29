package validate

import (
	"context"
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/hooks"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type conf struct {
	files               []string
	warcDir             string
	openOutputFileHook  hooks.OpenOutputFileHook
	closeOutputFileHook hooks.CloseOutputFileHook
}

var config conf

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "validate <files/dirs>",
		Short:             "Validate warc files",
		Long:              ``,
		RunE:              parseArgumentsAndCallValidate,
		ValidArgsFunction: flag.SuffixCompletionFn,
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	cmd.Flags().BoolP(flag.Recursive, "r", false, flag.RecursiveHelp)
	cmd.Flags().BoolP(flag.FollowSymlinks, "s", false, flag.FollowSymlinksHelp)
	cmd.Flags().StringSlice(flag.Suffixes, []string{".warc", ".warc.gz"}, flag.SuffixesHelp)
	cmd.Flags().IntP(flag.Concurrency, "c", int(float32(runtime.NumCPU())*float32(1.5)), flag.ConcurrencyHelp)
	cmd.Flags().String(flag.SrcFilesystem, "", flag.SrcFilesystemHelp)
	cmd.Flags().String(flag.SrcFileList, "", flag.SrcFileListHelp)
	cmd.Flags().BoolP(flag.KeepIndex, "k", false, flag.KeepIndexHelp)
	cmd.Flags().BoolP(flag.NewIndex, "K", false, flag.NewIndexHelp)
	cmd.Flags().StringP(flag.IndexDir, "i", cacheDir+"/warc", flag.IndexDirHelp)
	cmd.Flags().String(flag.OpenInputFileHook, "", flag.OpenInputFileHookHelp)
	cmd.Flags().String(flag.CloseInputFileHook, "", flag.CloseInputFileHookHelp)
	cmd.Flags().String(flag.OpenOutputFileHook, "", flag.OpenOutputFileHookHelp)
	cmd.Flags().String(flag.CloseOutputFileHook, "", flag.CloseOutputFileHookHelp)
	cmd.Flags().String(flag.WarcDir, "", "output directory for validated warc files. If not empty this enables copying of input file. Directory must exist.")
	cmd.Flags().String(flag.CalculateHash, "", flag.CalculateHashHelp)

	return cmd
}

func parseArgumentsAndCallValidate(cmd *cobra.Command, args []string) error {

	if len(args) == 0 && viper.GetString(flag.SrcFileList) == "" {
		return errors.New("missing file or directory name")
	}
	config.files = args
	config.warcDir = viper.GetString(flag.WarcDir)

	var err error
	config.openOutputFileHook, err = hooks.NewOpenOutputFileHook(cmd.Name(), viper.GetString(flag.OpenOutputFileHook))
	if err != nil {
		return err
	}

	config.closeOutputFileHook, err = hooks.NewCloseOutputFileHook(cmd.Name(), viper.GetString(flag.CloseOutputFileHook))
	if err != nil {
		return err
	}

	return runE(cmd.Name())
}

func runE(cmd string) error {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	fileWalker, err := filewalker.NewFromViper(cmd, config.files, validateFile)
	if err != nil {
		return err
	}
	stats := filewalker.NewStats()
	return fileWalker.Walk(ctx, stats)
}

func validateFile(fs afero.Fs, file string) filewalker.Result {
	result := filewalker.NewResult(file)
	var warcInfoId string

	r, err := newTeeReader(fs, file)
	if err != nil {
		result.AddError(err)
		return result
	}

	defer func() {
		_ = r.Close(&warcInfoId, result)
	}()

	wf, err := gowarc.NewWarcFileReaderFromStream(r, 0, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = wf.Close() }()

	for {
		wr, currentOffset, validation, err := wf.Next()
		if err == io.EOF {
			break
		}
		if wr.Type() == gowarc.Warcinfo {
			warcInfoId = wr.WarcHeader().GetId(gowarc.WarcRecordID)
		}
		result.IncrRecords()
		result.IncrProcessed()
		if err != nil {
			result.AddError(fmt.Errorf("rec num: %d, offset: %d, cause: %w", result.Records(), currentOffset, err))
			break
		}

		err = wr.ValidateDigest(validation)
		if err != nil {
			result.AddError(fmt.Errorf("rec num: %d, offset: %d, cause: %w", result.Records(), currentOffset, err))
			break
		}

		if err := wr.Close(); err != nil {
			*validation = append(*validation, err)
		}

		if !validation.Valid() {
			result.AddError(fmt.Errorf("rec num: %d, offset: %d, record: %s, cause: %w", result.Records(), currentOffset, wr, validation))
		}
	}
	return result
}

type teeReader struct {
	io.Reader
	r     afero.File
	w     *os.File
	rName string
	wName string
}

func newTeeReader(fs afero.Fs, file string) (*teeReader, error) {
	f, err := fs.Open(file)
	if err != nil {
		return nil, err
	}
	t := &teeReader{
		r:     f,
		rName: file,
	}

	if config.warcDir != "" {
		t.wName = config.warcDir + "/" + path.Base(file)
		if err := config.openOutputFileHook.WithSrcFileName(file).Run(t.wName); err != nil {
			_ = f.Close()
			return nil, err
		}
		of, err := os.Create(t.wName)
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		t.w = of
		t.Reader = io.TeeReader(f, t.w)
	} else {
		t.Reader = f
	}
	t.Reader = CountingReader(t.Reader)
	return t, nil
}

func (t *teeReader) Close(warcInfoId *string, result filewalker.Result) (err error) {
	if t.r != nil {
		_ = t.r.Close()
		t.r = nil
	}
	if t.w != nil {
		_ = t.w.Close()

		if err := config.closeOutputFileHook.
			WithSrcFileName(t.rName).
			WithHash(t.Hash()).
			WithErrorCount(result.ErrorCount()).
			Run(t.wName, t.Size(), *warcInfoId); err != nil {
			panic(err)
		}
	}

	return
}

func (t *teeReader) Size() int64 {
	if r, ok := t.Reader.(*countingReader); ok {
		return r.size
	}
	return 0
}

func (t *teeReader) Hash() string {
	if r, ok := t.Reader.(*countingReader); ok && r.hash != nil {
		return fmt.Sprintf("%0x", r.hash.Sum(nil))
	}
	return ""
}

func CountingReader(r io.Reader) io.Reader {
	c := &countingReader{Reader: r}
	switch viper.GetString(flag.CalculateHash) {
	case "md5":
		c.hash = crypto.MD5.New()
	case "sha1":
		c.hash = crypto.SHA1.New()
	case "sha256":
		c.hash = crypto.SHA256.New()
	case "sha512":
		c.hash = crypto.SHA512.New()
	}
	return c
}

type countingReader struct {
	io.Reader
	size int64
	hash hash.Hash
}

func (c *countingReader) Read(p []byte) (n int, err error) {
	n, err = c.Reader.Read(p)
	c.size += int64(n)
	if c.hash != nil {
		c.hash.Write(p[:n])
	}
	return
}
