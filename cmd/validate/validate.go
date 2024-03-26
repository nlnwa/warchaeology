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

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChannel
		cancel()
	}()

	fileWalker, err := filewalker.NewFromViper(cmd, config.files, validateFile)
	if err != nil {
		return err
	}
	stats := filewalker.NewStats()
	return fileWalker.Walk(ctx, stats)
}

func validateFile(fileSystem afero.Fs, file string) filewalker.Result {
	result := filewalker.NewResult(file)
	var warcInfoId string

	teeReader, err := newTeeReader(fileSystem, file)
	if err != nil {
		result.AddError(err)
		return result
	}

	defer func() {
		_ = teeReader.Close(&warcInfoId, result)
	}()

	warcFileReader, err := gowarc.NewWarcFileReaderFromStream(teeReader, 0, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = warcFileReader.Close() }()

	for {
		warcRecord, currentOffset, validation, err := warcFileReader.Next()
		if err == io.EOF {
			break
		}
		if warcRecord.Type() == gowarc.Warcinfo {
			warcInfoId = warcRecord.WarcHeader().GetId(gowarc.WarcRecordID)
		}
		result.IncrRecords()
		result.IncrProcessed()
		if err != nil {
			result.AddError(fmt.Errorf("rec num: %d, offset: %d, cause: %w", result.Records(), currentOffset, err))
			break
		}

		err = warcRecord.ValidateDigest(validation)
		if err != nil {
			result.AddError(fmt.Errorf("rec num: %d, offset: %d, cause: %w", result.Records(), currentOffset, err))
			break
		}

		if err := warcRecord.Close(); err != nil {
			*validation = append(*validation, err)
		}

		if !validation.Valid() {
			result.AddError(fmt.Errorf("rec num: %d, offset: %d, record: %s, cause: %w", result.Records(), currentOffset, warcRecord, validation))
		}
	}
	return result
}

type teeReader struct {
	io.Reader
	r              afero.File
	w              *os.File
	inputFileName  string
	outputFileName string
}

func newTeeReader(fs afero.Fs, file string) (*teeReader, error) {
	f, err := fs.Open(file)
	if err != nil {
		return nil, err
	}
	teeReader := &teeReader{
		r:             f,
		inputFileName: file,
	}

	if config.warcDir != "" {
		teeReader.outputFileName = config.warcDir + "/" + path.Base(file)
		if err := config.openOutputFileHook.WithSrcFileName(file).Run(teeReader.outputFileName); err != nil {
			_ = f.Close()
			return nil, err
		}
		file, err := os.Create(teeReader.outputFileName)
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		teeReader.w = file
		teeReader.Reader = io.TeeReader(f, teeReader.w)
	} else {
		teeReader.Reader = f
	}
	teeReader.Reader = CountingReader(teeReader.Reader)
	return teeReader, nil
}

func (reader *teeReader) Close(warcInfoId *string, result filewalker.Result) (err error) {
	if reader.r != nil {
		_ = reader.r.Close()
		reader.r = nil
	}
	if reader.w != nil {
		_ = reader.w.Close()

		if err := config.closeOutputFileHook.
			WithSrcFileName(reader.inputFileName).
			WithHash(reader.Hash()).
			WithErrorCount(result.ErrorCount()).
			Run(reader.outputFileName, reader.Size(), *warcInfoId); err != nil {
			panic(err)
		}
	}

	return
}

func (reader *teeReader) Size() int64 {
	if countingReader, ok := reader.Reader.(*countingReader); ok {
		return countingReader.size
	}
	return 0
}

func (reader *teeReader) Hash() string {
	if countingReader, ok := reader.Reader.(*countingReader); ok && countingReader.hash != nil {
		return fmt.Sprintf("%0x", countingReader.hash.Sum(nil))
	}
	return ""
}

func CountingReader(ioReader io.Reader) io.Reader {
	countingReader := &countingReader{Reader: ioReader}
	switch viper.GetString(flag.CalculateHash) {
	case "md5":
		countingReader.hash = crypto.MD5.New()
	case "sha1":
		countingReader.hash = crypto.SHA1.New()
	case "sha256":
		countingReader.hash = crypto.SHA256.New()
	case "sha512":
		countingReader.hash = crypto.SHA512.New()
	}
	return countingReader
}

type countingReader struct {
	io.Reader
	size int64
	hash hash.Hash
}

func (reader *countingReader) Read(byteSlice []byte) (length int, err error) {
	length, err = reader.Reader.Read(byteSlice)
	reader.size += int64(length)
	if reader.hash != nil {
		reader.hash.Write(byteSlice[:length])
	}
	return
}
