package warc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v3/cmd/internal/flag"
	"github.com/nlnwa/warchaeology/v3/internal/filewalker"
	"github.com/nlnwa/warchaeology/v3/internal/hooks"
	"github.com/nlnwa/warchaeology/v3/internal/index"
	"github.com/nlnwa/warchaeology/v3/internal/stat"
	"github.com/nlnwa/warchaeology/v3/internal/util"
	"github.com/nlnwa/warchaeology/v3/internal/warc"
	"github.com/nlnwa/warchaeology/v3/internal/warcwriterconfig"
	"github.com/nlnwa/warchaeology/v3/internal/workerpool"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ConvertWarcOptions struct {
	Paths              []string
	Concurrency        int
	MinWARCDiskFree    int64
	Repair             bool
	ContinueOnError    bool
	Offset             int64
	RecordNum          int
	RecordCount        int
	Force              bool
	WarcRecordOptions  []gowarc.WarcRecordOption
	WarcWriterConfig   *warcwriterconfig.WarcWriterConfig
	FileWalker         *filewalker.FileWalker
	FileIndex          *index.FileIndex
	OpenInputFileHook  hooks.OpenInputFileHook
	CloseInputFileHook hooks.CloseInputFileHook
}

type ConvertWarcFlags struct {
	FileWalkerFlags       flag.FileWalkerFlags
	IndexFlags            flag.IndexFlags
	WarcRecordOptionFlags flag.WarcRecordOptionFlags
	WarcWriterConfigFlags *flag.WarcWriterConfigFlags
	OutputHookFlags       *flag.OutputHookFlags
	InputHookFlags        *flag.InputHookFlags
	WarcIteratorFlags     flag.WarcIteratorFlags
	UtilFlags             flag.UtilFlags
	RepairFlags           flag.RepairFlags
	ConcurrencyFlags      flag.ConcurrencyFlags
	ErrorFlags            flag.ErrorFlags
}

func NewConvertWarcFlags() ConvertWarcFlags {
	return ConvertWarcFlags{
		OutputHookFlags:       &flag.OutputHookFlags{},
		InputHookFlags:        &flag.InputHookFlags{},
		WarcWriterConfigFlags: &flag.WarcWriterConfigFlags{},
	}
}

func (f ConvertWarcFlags) AddFlags(cmd *cobra.Command) {
	f.FileWalkerFlags.AddFlags(cmd, flag.WithDefaultSuffixes([]string{".warc", ".warc.gz"}))
	f.IndexFlags.AddFlags(cmd)
	f.WarcRecordOptionFlags.AddFlags(cmd)
	f.WarcWriterConfigFlags.AddFlags(cmd)
	f.OutputHookFlags.AddFlags(cmd)
	f.InputHookFlags.AddFlags(cmd)
	f.UtilFlags.AddFlags(cmd)
	f.RepairFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)
	f.ErrorFlags.AddFlags(cmd)
	f.WarcIteratorFlags.AddFlags(cmd)
}

func (f ConvertWarcFlags) ToConvertWarcOptions() (*ConvertWarcOptions, error) {
	wwc, err := f.WarcWriterConfigFlags.ToWarcWriterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create warc writer config: %w", err)
	}
	warcRecordOptions := []gowarc.WarcRecordOption{
		gowarc.WithVersion(wwc.WarcVersion),
		gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)),
	}
	if f.RepairFlags.Repair() {
		warcRecordOptions = append(warcRecordOptions,
			gowarc.WithSyntaxErrorPolicy(gowarc.ErrWarn),
			gowarc.WithSpecViolationPolicy(gowarc.ErrWarn),
			gowarc.WithAddMissingDigest(true),
			gowarc.WithFixSyntaxErrors(true),
			gowarc.WithFixDigest(true),
			gowarc.WithAddMissingContentLength(true),
			gowarc.WithAddMissingRecordId(true),
			gowarc.WithFixContentLength(true),
			gowarc.WithFixWarcFieldsBlockErrors(true),
		)
	} else {
		warcRecordOptions = append(warcRecordOptions,
			gowarc.WithSyntaxErrorPolicy(gowarc.ErrWarn),
			gowarc.WithSpecViolationPolicy(gowarc.ErrWarn),
			gowarc.WithAddMissingDigest(false),
			gowarc.WithFixSyntaxErrors(false),
			gowarc.WithFixDigest(false),
			gowarc.WithAddMissingContentLength(false),
			gowarc.WithAddMissingRecordId(false),
			gowarc.WithFixContentLength(false),
		)
	}

	fileWalker, err := f.FileWalkerFlags.ToFileWalker()
	if err != nil {
		return nil, fmt.Errorf("failed to create file walker: %w", err)
	}

	fileList, err := flag.ReadSrcFileList(f.FileWalkerFlags.SrcFileListFlags.SrcFileList())
	if err != nil {
		return nil, fmt.Errorf("failed to read from source file: %w", err)
	}

	openInputFileHook, err := f.InputHookFlags.ToOpenInputFileHook()
	if err != nil {
		return nil, fmt.Errorf("failed to create open input file hook: %w", err)
	}

	closeInputFileHook, err := f.InputHookFlags.ToCloseInputFileHook()
	if err != nil {
		return nil, fmt.Errorf("failed to create close input file hook: %w", err)
	}

	var fileIndex *index.FileIndex
	if f.IndexFlags.KeepIndex() {
		fileIndex, err = f.IndexFlags.ToFileIndex()
		if err != nil {
			return nil, fmt.Errorf("failed to create file index: %w", err)
		}
	}

	return &ConvertWarcOptions{
		Concurrency:        f.ConcurrencyFlags.Concurrency(),
		MinWARCDiskFree:    f.UtilFlags.MinFreeDisk(),
		WarcWriterConfig:   wwc,
		FileWalker:         fileWalker,
		WarcRecordOptions:  warcRecordOptions,
		Paths:              fileList,
		FileIndex:          fileIndex,
		RecordNum:          f.WarcIteratorFlags.RecordNum(),
		RecordCount:        f.WarcIteratorFlags.Limit(),
		Force:              f.WarcIteratorFlags.Force(),
		Offset:             f.WarcIteratorFlags.Offset(),
		OpenInputFileHook:  openInputFileHook,
		CloseInputFileHook: closeInputFileHook,
		ContinueOnError:    f.ErrorFlags.ContinueOnError(),
	}, nil
}

func NewCmdConvertWarc() *cobra.Command {
	flags := NewConvertWarcFlags()

	var cmd = &cobra.Command{
		Use:   "warc FILE/DIR ...",
		Short: "Convert WARC file into WARC file",
		Long: `The WARC to WARC converter can be used to reorganize, convert or repair WARC-records.
This is an experimental feature.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToConvertWarcOptions()
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

func (o *ConvertWarcOptions) Complete(cmd *cobra.Command, args []string) error {
	o.Paths = append(o.Paths, args...)
	return nil
}

func (o *ConvertWarcOptions) Validate() error {
	if len(o.Paths) == 0 {
		return errors.New("missing file or directory name")
	}
	return nil
}

func (o *ConvertWarcOptions) Run() error {
	done := make(chan struct{})
	exitCode := 0

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
			slog.Info("Converted file", "errors", result.ErrorCount(), "records", result.Records())
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
	defer o.WarcWriterConfig.Close()

	workerPool := workerpool.New(ctx, o.Concurrency)
	defer workerPool.CloseWait()

	for _, path := range o.Paths {
		err := o.FileWalker.Walk(ctx, path, func(fs afero.Fs, path string, err error) error {
			if err != nil {
				return err
			}

			workerPool.Submit(func() {
				// Assert WARC disk has enough free space
				if o.MinWARCDiskFree > 0 {
					diskFree, err := util.DiskFree(o.WarcWriterConfig.OutDir)
					if err != nil {
						cancel()
						slog.Error("Failed to get free space on device", "path", o.WarcWriterConfig.OutDir, "error", err)
						return
					}
					if diskFree < o.MinWARCDiskFree {
						cancel()
						slog.Error("Not enough free space on device", "bytesFree", diskFree, "path", o.WarcWriterConfig.OutDir)
						return
					}
				}

				result, err := filewalker.Preposterous(fs, path, o.OpenInputFileHook, o.CloseInputFileHook, o.FileIndex, o.handleFile)
				if errors.Is(err, filewalker.ErrSkipFile) {
					return
				}
				if err != nil {
					if !o.ContinueOnError {
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

func (o *ConvertWarcOptions) handleFile(fileSystem afero.Fs, path string) (stat.Result, error) {
	file, err := fileSystem.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	warcFileReader, err := gowarc.NewWarcFileReaderFromStream(file, o.Offset, o.WarcRecordOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create warc file reader: %w", err)
	}
	defer func() { _ = warcFileReader.Close() }()

	var writer *gowarc.WarcFileWriter
	if o.WarcWriterConfig.OneToOneWriter {
		defer func() {
			if writer != nil {
				_ = writer.Close()
			}
		}()
	}

	result := stat.NewResult(path)

	var lastOffset int64 = -1

	for record, err := range warc.Records(warcFileReader, nil, o.RecordNum, o.RecordCount) {
		if err != nil {
			// When forcing, avoid infinite loop by ensuring the iterator moves forward
			if o.Force && lastOffset != record.Offset {
				slog.Warn(err.Error(), "offset", record.Offset, "path", path)
				lastOffset = record.Offset
				continue
			}
			return result, warc.Error(record, err)
		}

		if writer == nil || !o.WarcWriterConfig.OneToOneWriter {
			writer, err = o.WarcWriterConfig.GetWarcWriter(path, record.WarcRecord.WarcHeader().Get(gowarc.WarcDate))
			if err != nil {
				return result, warc.Error(record, err)
			}
		}
		err = o.handleRecord(writer, record, result)
		if err != nil {
			return result, warc.Error(record, err)
		}
	}
	return result, nil
}

func (o *ConvertWarcOptions) handleRecord(warcFileWriter *gowarc.WarcFileWriter, record warc.Record, result stat.Result) error {
	defer record.Close()

	result.IncrRecords()

	if !record.Validation.Valid() {
		for _, err := range *record.Validation {
			result.AddError(warc.Error(record, err))
		}
	}

	warcRecord := record.WarcRecord
	warcRecordBuilder := gowarc.NewRecordBuilder(
		warcRecord.Type(),
		gowarc.WithFixContentLength(false),
		gowarc.WithFixDigest(false),
	)
	for _, warcField := range *warcRecord.WarcHeader() {
		if warcField.Name != gowarc.WarcType {
			warcRecordBuilder.AddWarcHeader(warcField.Name, warcField.Value)
		}
	}
	ioReader, err := warcRecord.Block().RawBytes()
	if err != nil {
		return err
	}
	_, err = warcRecordBuilder.ReadFrom(ioReader)
	if err != nil {
		return err
	}
	warcRecord, _, err = warcRecordBuilder.Build()
	if err != nil {
		return err
	}
	if writeResponse := warcFileWriter.Write(warcRecord); len(writeResponse) > 0 {
		return writeResponse[0].Err
	}
	return nil
}
