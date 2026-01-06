package arc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nationallibraryofnorway/warchaeology/v4/arcreader"
	"github.com/nationallibraryofnorway/warchaeology/v4/cmd/internal/flag"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/filewalker"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/hooks"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/index"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/stat"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/util"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/version"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/warc"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/warcwriterconfig"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/workerpool"
	"github.com/nlnwa/gowarc/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ConvertArcOptions struct {
	Paths              []string
	Concurrency        int
	MinWARCDiskFree    int64
	WarcWriterConfig   *warcwriterconfig.WarcWriterConfig
	WarcRecordOptions  []gowarc.WarcRecordOption
	FileWalker         *filewalker.FileWalker
	ContinueOnError    bool
	FileIndex          *index.FileIndex
	OpenInputFileHook  hooks.OpenInputFileHook
	CloseInputFileHook hooks.CloseInputFileHook
}

type ConvertArcFlags struct {
	FileWalkerFlags       flag.FileWalkerFlags
	IndexFlags            flag.IndexFlags
	WarcWriterConfigFlags *flag.WarcWriterConfigFlags
	WarcRecordOptionFlags flag.WarcRecordOptionFlags
	OutputHookFlags       *flag.OutputHookFlags
	InputHookFlags        *flag.InputHookFlags
	UtilFlags             flag.UtilFlags
	ConcurrencyFlags      flag.ConcurrencyFlags
	ErrorFlags            flag.ErrorFlags
}

func NewConvertArcFlags() ConvertArcFlags {
	return ConvertArcFlags{
		OutputHookFlags:       &flag.OutputHookFlags{},
		InputHookFlags:        &flag.InputHookFlags{},
		WarcWriterConfigFlags: &flag.WarcWriterConfigFlags{},
	}
}

func (f ConvertArcFlags) AddFlags(cmd *cobra.Command) {
	f.FileWalkerFlags.AddFlags(cmd, flag.WithDefaultSuffixes([]string{".arc", ".arc.gz"}))
	f.OutputHookFlags.AddFlags(cmd)
	f.InputHookFlags.AddFlags(cmd)
	f.WarcWriterConfigFlags.AddFlags(cmd, flag.WithDefaultOneToOne(true))
	f.WarcRecordOptionFlags.AddFlags(cmd)
	f.IndexFlags.AddFlags(cmd)
	f.UtilFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)
	f.ErrorFlags.AddFlags(cmd)
}

func (f ConvertArcFlags) ToConvertArcOptions() (*ConvertArcOptions, error) {
	warcWriterConfig, err := f.WarcWriterConfigFlags.ToWarcWriterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create warc writer config: %w", err)
	}
	if warcWriterConfig.OneToOneWriter {
		warcInfoFunc := func(recordBuilder gowarc.WarcRecordBuilder) error {
			payload := &gowarc.WarcFields{}
			payload.Set("software", version.SoftwareVersion())
			payload.Set("format", fmt.Sprintf("WARC File Format %d.%d", warcWriterConfig.WarcVersion.Minor(), warcWriterConfig.WarcVersion.Minor()))
			payload.Set("description", "Converted from ARC")
			hostname, errInner := os.Hostname()
			if errInner != nil {
				return errInner
			}
			payload.Set("host", hostname)

			_, err := recordBuilder.WriteString(payload.String())
			return err
		}
		warcWriterConfig.WarcFileWriterOptions = append(warcWriterConfig.WarcFileWriterOptions, gowarc.WithWarcInfoFunc(warcInfoFunc))
	}

	warcRecordOptions := f.WarcRecordOptionFlags.ToWarcRecordOptions()
	warcRecordOptions = append(warcRecordOptions,
		gowarc.WithVersion(warcWriterConfig.WarcVersion),
		gowarc.WithAddMissingDigest(true),
	)

	openInputFileHook, err := f.InputHookFlags.ToOpenInputFileHook()
	if err != nil {
		return nil, fmt.Errorf("failed to create open input file hook: %w", err)
	}

	closeInputFileHook, err := f.InputHookFlags.ToCloseInputFileHook()
	if err != nil {
		return nil, fmt.Errorf("failed to create close input file hook: %w", err)
	}

	// and we also read paths from a file if the --src-file-fileList flag is set
	fileList, err := flag.ReadSrcFileList(viper.GetString(flag.SrcFileList))
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

	return &ConvertArcOptions{
		Concurrency:        f.ConcurrencyFlags.Concurrency(),
		OpenInputFileHook:  openInputFileHook,
		CloseInputFileHook: closeInputFileHook,
		WarcWriterConfig:   warcWriterConfig,
		WarcRecordOptions:  warcRecordOptions,
		FileWalker:         fileWalker,
		FileIndex:          fileIndex,
		Paths:              fileList,
		ContinueOnError:    f.ErrorFlags.ContinueOnError(),
		MinWARCDiskFree:    f.UtilFlags.MinFreeDisk(),
	}, nil
}

func NewCmdConvertArc() *cobra.Command {
	flags := NewConvertArcFlags()

	var cmd = &cobra.Command{
		Use:   "arc FILE/DIR ...",
		Short: "Convert ARC to WARC",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToConvertArcOptions()
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

func (o *ConvertArcOptions) Complete(cmd *cobra.Command, args []string) error {
	o.Paths = append(o.Paths, args...)
	return nil
}

func (o *ConvertArcOptions) Validate() error {
	if len(o.Paths) == 0 {
		return errors.New("missing file or directory name")
	}
	return nil
}

func (o *ConvertArcOptions) Run() error {
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
				} else if err != nil {
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

func (o *ConvertArcOptions) handleFile(fs afero.Fs, fileName string) (stat.Result, error) {
	result := stat.NewResult(fileName)

	arcFileReader, err := arcreader.NewArcFileReader(fs, fileName, 0, o.WarcRecordOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create arc file reader: %w", err)
	}
	defer func() { _ = arcFileReader.Close() }()

	var writer *gowarc.WarcFileWriter
	if o.WarcWriterConfig.OneToOneWriter {
		defer func() {
			if writer != nil {
				_ = writer.Close()
			}
		}()
	}

	for record, err := range warc.Records(arcFileReader, nil, 0, 0) {
		if err != nil {
			return result, warc.Error(record, err)
		}
		if writer == nil || !o.WarcWriterConfig.OneToOneWriter {
			warcDate, err := record.WarcRecord.WarcHeader().GetTime(gowarc.WarcDate)
			if err != nil {
				return result, warc.Error(record, err)
			}
			writer, err = o.WarcWriterConfig.GetWarcWriter(fileName, warcDate)
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

func (o *ConvertArcOptions) handleRecord(warcFileWriter *gowarc.WarcFileWriter, record warc.Record, result stat.Result) error {
	defer record.Close()

	result.IncrRecords()

	if !record.Validation.Valid() {
		for _, err := range *record.Validation {
			result.AddError(warc.Error(record, err))
		}
	}

	writeResponse := warcFileWriter.Write(record.WarcRecord)
	if len(writeResponse) > 0 {
		return writeResponse[0].Err
	}
	return nil
}
