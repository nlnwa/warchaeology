package validate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v3/cmd/internal/flag"
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
	offset              int64
	force               bool
	hashFunction        string
	FileIndex           *index.FileIndex
	filter              *filter.RecordFilter
	FileWalker          *filewalker.FileWalker
	continueOnError     bool
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
	ErrorFlags            flag.ErrorFlags
	WarcIteratorFlags     flag.WarcIteratorFlags
}

func (f ValidateFlags) AddFlags(cmd *cobra.Command) {
	f.FileWalkerFlags.AddFlags(cmd)
	f.FilterFlags.AddFlags(cmd)
	f.IndexFlags.AddFlags(cmd)
	f.WarcRecordOptionFlags.AddFlags(cmd)
	f.OutputHookFlags.AddFlags(cmd)
	f.InputHookFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)
	f.ErrorFlags.AddFlags(cmd)
	f.WarcIteratorFlags.AddFlags(cmd)

	cmd.Flags().String(CalculateHash, "", CalculateHashHelp)
	cmd.Flags().StringP(flag.OutputDir, "w", "", "output directory for validated warc files. If not empty this enables copying of input file. Directory must exist.")
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
		continueOnError:     f.ErrorFlags.ContinueOnError(),
		filter:              filter,
		hashFunction:        f.HashFunction(),
		concurrency:         f.ConcurrencyFlags.Concurrency(),
		outputDir:           f.OutputDir(),
		recordNum:           f.WarcIteratorFlags.RecordNum(),
		recordCount:         f.WarcIteratorFlags.Limit(),
		force:               f.WarcIteratorFlags.Force(),
		offset:              f.WarcIteratorFlags.Offset(),
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
			err = o.Run()
			if errors.Is(err, context.Canceled) {
				os.Exit(1)
			}
			return err
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
	exitCode := 0
	done := make(chan struct{})

	defer func() {
		<-done
		os.Exit(exitCode)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	results := make(chan stat.Result)
	go func() {
		defer close(done)

		stats := stat.NewStats()
		defer func() {
			slog.Info("Total", "files", stats.Files, "errors", stats.Errors, "records", stats.Records)
		}()

		for result := range results {
			slog := slog.With("path", result.Name())
			for _, err := range result.Errors() {
				var recordErr warc.RecordError
				if errors.As(err, &recordErr) {
					slog.Error("Validation error", "error", recordErr.Error(), "offset", recordErr.Offset())
				} else {
					slog.Error("Validation error", "error", err.Error())
				}
			}
			slog.Info("Validated file", "errors", result.ErrorCount(), "records", result.Records())
			stats.Merge(result)
		}
		if stats.Errors > 0 {
			exitCode = 1
		}
	}()
	defer close(results)

	if o.FileIndex != nil {
		defer o.FileIndex.Close()
	}

	workerPool := workerpool.New(ctx, o.concurrency)
	defer workerPool.CloseWait()

	for _, path := range o.paths {
		err := o.FileWalker.Walk(ctx, path, func(fs afero.Fs, path string, err error) error {
			if err != nil {
				return err
			}

			workerPool.Submit(func() {
				result, err := filewalker.Preposterous(fs, path, o.openInputFileHook, o.closeInputFileHook, o.FileIndex, o.handleFile)
				if errors.Is(err, filewalker.ErrSkipFile) {
					return
				} else if err != nil {
					if !o.continueOnError {
						cancel()
					}
					if result == nil {
						result = stat.NewResult(path)
					}
					result.AddError(err)
				}

				results <- result
			})

			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *ValidateOptions) handleFile(fs afero.Fs, path string) (stat.Result, error) {
	result := stat.NewResult(path)
	var warcInfoId string

	teeReader, err := o.newTeeReader(fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create tee reader: %w", err)

	}
	defer func() { _ = teeReader.Close(&warcInfoId, result) }()

	warcFileReader, err := gowarc.NewWarcFileReaderFromStream(teeReader, o.offset, o.warcRecordOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create warc file reader: %w", err)
	}
	defer func() { _ = warcFileReader.Close() }()

	var lastOffset int64 = -1

	for record, err := range warc.Records(warcFileReader, o.filter, o.recordNum, o.recordCount) {
		if err != nil {
			// When forcing, avoid infinite loop by ensuring the iterator moves forward
			if o.force && lastOffset != record.Offset {
				slog.Warn(err.Error(), "offset", record.Offset, "path", path)
				lastOffset = record.Offset
				continue
			}
			return result, warc.Error(record, err)
		}
		// Capture the WARC Record Id of the first warcinfo record in the file.
		// It is passed to the close output file hook as an environment variable
		if warcInfoId == "" && record.WarcRecord.Type() == gowarc.Warcinfo {
			warcInfoId = record.WarcRecord.WarcHeader().GetId(gowarc.WarcRecordID)
		}
		err = handleRecord(record, result)
		if err != nil {
			return result, warc.Error(record, err)
		}
	}
	return result, nil
}

func handleRecord(record warc.Record, result stat.Result) error {
	defer record.Close()

	result.IncrRecords()

	err := record.WarcRecord.ValidateDigest(record.Validation)
	for _, err := range *record.Validation {
		result.AddError(err)
	}
	return err
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
