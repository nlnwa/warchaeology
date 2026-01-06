package validate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nationallibraryofnorway/warchaeology/v4/cmd/internal/flag"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/filewalker"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/filter"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/hooks"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/index"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/stat"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/warc"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/workerpool"
	"github.com/nlnwa/gowarc/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	CalculateHash     = "calculate-hash"
	CalculateHashHelp = `calculate hash of output file. The hash is made available to the close output file hook as WARC_HASH. Valid values: md5, sha1, sha256, sha512`
)

type ValidateOptions struct {
	paths              []string
	outputDir          string
	concurrency        int
	recordNum          int
	recordCount        int
	offset             int64
	force              bool
	hashFunction       string
	FileIndex          *index.FileIndex
	filter             *filter.RecordFilter
	FileWalker         *filewalker.FileWalker
	continueOnError    bool
	warcRecordOptions  []gowarc.WarcRecordOption
	openInputFileHook  hooks.OpenInputFileHook
	closeInputFileHook hooks.CloseInputFileHook
}

func NewValidateFlags() ValidateFlags {
	return ValidateFlags{
		OutputHookFlags: &flag.OutputHookFlags{},
		InputHookFlags:  &flag.InputHookFlags{},
	}
}

type ValidateFlags struct {
	IndexFlags            flag.IndexFlags
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
	f.InputHookFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)
	f.ErrorFlags.AddFlags(cmd)
	f.WarcIteratorFlags.AddFlags(cmd)

	cmd.Flags().String(CalculateHash, "", CalculateHashHelp)
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

	return &ValidateOptions{
		paths:              fileList,
		FileWalker:         fileWalker,
		continueOnError:    f.ErrorFlags.ContinueOnError(),
		filter:             filter,
		hashFunction:       f.HashFunction(),
		concurrency:        f.ConcurrencyFlags.Concurrency(),
		outputDir:          f.OutputDir(),
		recordNum:          f.WarcIteratorFlags.RecordNum(),
		recordCount:        f.WarcIteratorFlags.Limit(),
		force:              f.WarcIteratorFlags.Force(),
		offset:             f.WarcIteratorFlags.Offset(),
		warcRecordOptions:  f.WarcRecordOptionFlags.ToWarcRecordOptions(),
		FileIndex:          fileIndex,
		openInputFileHook:  openInputFileHook,
		closeInputFileHook: closeInputFileHook,
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

	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	countingReader := NewCountingReader(file, o.hashFunction)
	defer func() {
		result.SetHash(countingReader.Hash())
	}()

	warcFileReader, err := gowarc.NewWarcFileReaderFromStream(countingReader, o.offset, o.warcRecordOptions...)
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

		func() {
			defer record.Close()

			result.IncrRecords()

			for _, err := range *record.Validation {
				result.AddError(warc.Error(record, err))
			}
		}()
	}
	return result, nil
}
