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
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/v3/cmd/internal/flag"
	"github.com/nlnwa/warchaeology/v3/cmd/internal/log"
	"github.com/nlnwa/warchaeology/v3/internal/filewalker"
	"github.com/nlnwa/warchaeology/v3/internal/filter"
	"github.com/nlnwa/warchaeology/v3/internal/hooks"
	"github.com/nlnwa/warchaeology/v3/internal/index"
	"github.com/nlnwa/warchaeology/v3/internal/stat"
	"github.com/nlnwa/warchaeology/v3/internal/warc"
	"github.com/nlnwa/warchaeology/v3/internal/workerpool"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	CalculateHash     = "calculate-hash"
	CalculateHashHelp = `calculate hash of output file. The hash is made available to the close output file hook as WARC_HASH. Valid values: md5, sha1, sha256, sha512`
)

type ValidateOptions struct {
	paths               []string
	outputDir           string
	concurrency         int
	recordNum           int
	recordCount         int
	hashFunction        string
	FileIndex           *index.FileIndex
	filter              *filter.RecordFilter
	FileWalker          *filewalker.FileWalker
	warcRecordOptions   []gowarc.WarcRecordOption
	openInputFileHook   hooks.OpenInputFileHook
	closeInputFileHook  hooks.CloseInputFileHook
	openOutputFileHook  hooks.OpenOutputFileHook
	closeOutputFileHook hooks.CloseOutputFileHook
}

func NewValidateFlags() ValidateFlags {
	return ValidateFlags{
		IndexFlags:      &flag.IndexFlags{},
		OutputHookFlags: &flag.OutputHookFlags{},
		InputHookFlags:  &flag.InputHookFlags{},
	}
}

type ValidateFlags struct {
	IndexFlags            *flag.IndexFlags
	OutputHookFlags       *flag.OutputHookFlags
	InputHookFlags        *flag.InputHookFlags
	FileWalkerFlags       flag.FileWalkerFlags
	FilterFlags           flag.FilterFlags
	WarcRecordOptionFlags flag.WarcRecordOptionFlags
	ConcurrencyFlags      flag.ConcurrencyFlags
}

func (f ValidateFlags) AddFlags(cmd *cobra.Command) {
	f.FileWalkerFlags.AddFlags(cmd)
	f.FilterFlags.AddFlags(cmd)
	f.IndexFlags.AddFlags(cmd)
	f.WarcRecordOptionFlags.AddFlags(cmd)
	f.OutputHookFlags.AddFlags(cmd)
	f.InputHookFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)

	cmd.Flags().String(CalculateHash, "", CalculateHashHelp)
	cmd.Flags().StringP(flag.OutputDir, "o", "", "output directory for validated warc files. If not empty this enables copying of input file. Directory must exist.")
}

func (f ValidateFlags) OutputDir() string {
	return viper.GetString(flag.OutputDir)
}

func (f ValidateFlags) HashFunction() string {
	return viper.GetString(CalculateHash)
}

func (f ValidateFlags) ToOptions() (*ValidateOptions, error) {
	filter, err := f.FilterFlags.ToFilter()
	if err != nil {
		return nil, fmt.Errorf("failed to create filter: %w", err)
	}
	fileList, err := flag.ReadSrcFileList(f.FileWalkerFlags.SrcFileListFlags.SrcFileList())
	if err != nil {
		return nil, fmt.Errorf("failed to read from source file: %w", err)
	}
	fileWalker, err := f.FileWalkerFlags.ToFileWalker()
	if err != nil {
		return nil, fmt.Errorf("failed to create file walker: %w", err)
	}
	var fileIndex *index.FileIndex
	if f.IndexFlags.KeepIndex() {
		fileIndex, err = f.IndexFlags.ToFileIndex()
		if err != nil {
			return nil, fmt.Errorf("failed to create file index: %w", err)
		}
	}
	openInputFileHook, err := f.InputHookFlags.ToOpenInputFileHook()
	if err != nil {
		return nil, fmt.Errorf("failed to create open input file hook: %w", err)
	}
	closeInputFileHook, err := f.InputHookFlags.ToCloseInputFileHook()
	if err != nil {
		return nil, fmt.Errorf("failed to create close input file hook: %w", err)
	}
	openOutputFileHook, err := hooks.NewOpenOutputFileHook("validate", f.OutputHookFlags.OpenOutputFileHook())
	if err != nil {
		return nil, fmt.Errorf("failed to create open output file hook: %w", err)
	}
	closeOutputFileHook, err := hooks.NewCloseOutputFileHook("validate", f.OutputHookFlags.CloseOutputFileHook())
	if err != nil {
		return nil, fmt.Errorf("failed to create close output file hook: %w", err)
	}

	return &ValidateOptions{
		paths:               fileList,
		FileWalker:          fileWalker,
		filter:              filter,
		hashFunction:        f.HashFunction(),
		concurrency:         f.ConcurrencyFlags.Concurrency(),
		outputDir:           f.OutputDir(),
		warcRecordOptions:   f.WarcRecordOptionFlags.ToWarcRecordOptions(),
		FileIndex:           fileIndex,
		openInputFileHook:   openInputFileHook,
		closeInputFileHook:  closeInputFileHook,
		openOutputFileHook:  openOutputFileHook,
		closeOutputFileHook: closeOutputFileHook,
	}, nil
}

func NewCmdValidate() *cobra.Command {
	flags := NewValidateFlags()

	var cmd = &cobra.Command{
		Use:   "validate FILE/DIR ...",
		Short: "Validate WARC files",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToOptions()
			if err != nil {
				return err
			}
			err = o.Complete(cmd, args)
			if err != nil {
				return err
			}
			err = o.Validate()
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return o.Run()
		},
		ValidArgsFunction: flag.SuffixCompletionFn,
	}

	flags.AddFlags(cmd)

	return cmd
}

func (o *ValidateOptions) Complete(cmd *cobra.Command, args []string) error {
	o.paths = append(o.paths, args...)
	return nil
}

func (o *ValidateOptions) Validate() error {
	if len(o.paths) == 0 {
		return errors.New("missing file or directory name")
	}
	return nil
}

func (o *ValidateOptions) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if o.FileIndex != nil {
		defer o.FileIndex.Close()
	}

	closer, err := log.InitLogger(os.Stderr)
	if err != nil {
		return err
	}
	defer closer.Close()

	workerPool := workerpool.New(o.concurrency)
	defer workerPool.CloseWait()

	walkFn := func(fs afero.Fs, path string, err error) error {
		if err != nil {
			return err
		}
		return workerPool.Submit(ctx, func() {
			result, err := filewalker.Preposterous(path, o.FileIndex, o.openInputFileHook, o.closeInputFileHook, func() stat.Result {
				return o.validateFile(ctx, fs, path)
			})
			if errors.Is(err, filewalker.ErrSkipFile) {
				return
			}
			if err != nil {
				slog.Error(err.Error())
				return
			}
			slog := slog.With("file", path, "errors", result.ErrorCount(), "records", result.Records())
			if result.ErrorCount() == 0 {
				slog.Info("Valid")
			} else {
				slog.Warn("Invalid")
			}
			if result.Fatal() != nil {
				defer cancel()
				slog.Error(result.Fatal().Error())
			}
			for _, err := range result.Errors() {
				slog.Error(err.Error())
			}
		})
	}

	for _, path := range o.paths {
		err := o.FileWalker.Walk(path, walkFn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *ValidateOptions) validateFile(ctx context.Context, fs afero.Fs, path string) stat.Result {
	result := stat.NewResult(path)
	var warcInfoId string

	teeReader, err := o.newTeeReader(fs, path)
	if err != nil {
		result.AddError(fmt.Errorf("failed to create tee reader: %w", err))
		return result
	}
	defer func() {
		_ = teeReader.Close(&warcInfoId, result)
	}()

	warcFileReader, err := gowarc.NewWarcFileReaderFromStream(teeReader, 0, o.warcRecordOptions...)
	if err != nil {
		result.AddError(fmt.Errorf("failed to create warc file reader: %w", err))
		return result
	}
	defer func() { _ = warcFileReader.Close() }()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var once sync.Once

	for record := range warc.NewIterator(ctx, warcFileReader, o.filter, o.recordNum, o.recordCount) {
		handleRecord(record, result)
		if result.Fatal() != nil {
			break
		}
		// Capture the WARC Record Id of the first warcinfo record in the file.
		// It is passed to the close output file hook as an environment variable
		if record.WarcRecord.Type() == gowarc.Warcinfo {
			once.Do(func() {
				warcInfoId = record.WarcRecord.WarcHeader().GetId(gowarc.WarcRecordID)
			})
		}
	}
	return result
}

func handleRecord(record warc.Record, result stat.Result) {
	defer record.Close()

	result.IncrRecords()
	result.IncrProcessed()

	if record.Err != nil {
		// a record iterator error is fatal because it is not possible to continue processing the file
		result.SetFatal(warc.Error(record, record.Err))
		return
	}
	if !record.Validation.Valid() {
		result.AddError(warc.Error(record, record.Validation))
	}
	if err := record.WarcRecord.ValidateDigest(record.Validation); err != nil {
		result.SetFatal(warc.Error(record, record.Err))
		return
	}
}

type teeReader struct {
	io.Reader
	r                   afero.File
	w                   *os.File
	inputFileName       string
	outputFileName      string
	closeOutputFileHook hooks.CloseOutputFileHook
}

func (o *ValidateOptions) newTeeReader(fs afero.Fs, file string) (*teeReader, error) {
	f, err := fs.Open(file)
	if err != nil {
		return nil, err
	}
	teeReader := &teeReader{
		r:                   f,
		inputFileName:       file,
		closeOutputFileHook: o.closeOutputFileHook,
	}

	if o.outputDir != "" {
		teeReader.outputFileName = filepath.Join(o.outputDir, filepath.Base(file))
		if err := o.openOutputFileHook.WithSrcFileName(file).Run(teeReader.outputFileName); err != nil {
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
	teeReader.Reader = NewCountingReader(teeReader.Reader, o.hashFunction)
	return teeReader, nil
}

func (reader *teeReader) Close(warcInfoId *string, result stat.Result) (err error) {
	if reader.r != nil {
		_ = reader.r.Close()
		reader.r = nil
	}
	if reader.w != nil {
		_ = reader.w.Close()

		return reader.closeOutputFileHook.
			WithSrcFileName(reader.inputFileName).
			WithHash(reader.Hash()).
			WithErrorCount(result.ErrorCount()).
			Run(reader.outputFileName, reader.Size(), *warcInfoId)
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

func NewCountingReader(ioReader io.Reader, hashFunction string) io.Reader {
	countingReader := &countingReader{Reader: ioReader}
	switch hashFunction {
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
