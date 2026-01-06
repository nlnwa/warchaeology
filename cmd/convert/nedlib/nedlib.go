package nedlib

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
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/hooks"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/index"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/stat"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/time"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/util"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/version"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/warc"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/warcwriterconfig"
	"github.com/nationallibraryofnorway/warchaeology/v4/internal/workerpool"
	"github.com/nationallibraryofnorway/warchaeology/v4/nedlibreader"
	"github.com/nlnwa/gowarc/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ConvertNedlibOptions struct {
	Paths              []string
	FileWalker         *filewalker.FileWalker
	WarcWriterConfig   *warcwriterconfig.WarcWriterConfig
	FileIndex          *index.FileIndex
	MinWARCDiskFree    int64
	Concurrency        int
	ContinueOnError    bool
	WarcRecordOptions  []gowarc.WarcRecordOption
	OpenInputFileHook  hooks.OpenInputFileHook
	CloseInputFileHook hooks.CloseInputFileHook
}

type ConvertNedlibFlags struct {
	FileWalkerFlags       flag.FileWalkerFlags
	WarcWriterConfigFlags *flag.WarcWriterConfigFlags
	IndexFlags            flag.IndexFlags
	OutputHookFlags       *flag.OutputHookFlags
	InputHookFlags        *flag.InputHookFlags
	ConcurrencyFlags      flag.ConcurrencyFlags
	ErrorFlags            flag.ErrorFlags
}

func NewConvertNedlibFlags() ConvertNedlibFlags {
	return ConvertNedlibFlags{
		OutputHookFlags:       &flag.OutputHookFlags{},
		InputHookFlags:        &flag.InputHookFlags{},
		WarcWriterConfigFlags: &flag.WarcWriterConfigFlags{},
	}
}

func (f ConvertNedlibFlags) AddFlags(cmd *cobra.Command) {
	f.FileWalkerFlags.AddFlags(cmd, flag.WithDefaultSuffixes([]string{".meta"}))
	f.WarcWriterConfigFlags.AddFlags(cmd, flag.WithDefaultFilePrefix("nedlib_"))
	f.IndexFlags.AddFlags(cmd)
	f.OutputHookFlags.AddFlags(cmd)
	f.InputHookFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)
	f.ErrorFlags.AddFlags(cmd)
}

func (f ConvertNedlibFlags) ToOptions() (*ConvertNedlibOptions, error) {
	// The Nedlib data structure does not support direct filename transformations.
	// Instead, we employ a custom generator that treats the input filename as a date.
	// When we request a new warcwriter, we submit a synthetic fromFilename based on the date of the first record.
	viper.Set(flag.NameGenerator, "nedlib")

	warcWriterConfig, err := f.WarcWriterConfigFlags.ToWarcWriterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create warc writer config: %w", err)
	}

	warcInfoFunc := func(wr gowarc.WarcRecordBuilder) error {
		warcHeader := &gowarc.WarcFields{}
		warcHeader.Set("software", version.SoftwareVersion())
		warcHeader.Set("description", "Converted from Nedlib")
		if hostname, err := os.Hostname(); err != nil {
			return err
		} else {
			warcHeader.Set("host", hostname)
		}
		_, err := wr.WriteString(warcHeader.String())
		return err
	}
	warcWriterConfig.WarcFileWriterOptions = append(warcWriterConfig.WarcFileWriterOptions, gowarc.WithWarcInfoFunc(warcInfoFunc))

	warcRecordOptions := []gowarc.WarcRecordOption{
		gowarc.WithVersion(warcWriterConfig.WarcVersion),
		gowarc.WithAddMissingDigest(true),
		gowarc.WithFixDigest(true),
		gowarc.WithFixContentLength(true),
		gowarc.WithAddMissingContentLength(true),
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

	return &ConvertNedlibOptions{
		WarcWriterConfig:  warcWriterConfig,
		WarcRecordOptions: warcRecordOptions,
		Paths:             fileList,
		FileWalker:        fileWalker,
		FileIndex:         fileIndex,
		ContinueOnError:   f.ErrorFlags.ContinueOnError(),
	}, nil
}

func NewCmdConvertNedlib() *cobra.Command {
	flags := NewConvertNedlibFlags()

	cmd := &cobra.Command{
		Use:   "nedlib FILE/DIR ...",
		Short: "Convert directory with files harvested with Nedlib into warc files",
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
	}

	flags.AddFlags(cmd)

	return cmd
}

func (o *ConvertNedlibOptions) Complete(cmd *cobra.Command, args []string) error {
	o.Paths = append(o.Paths, args...)
	return nil
}

func (o *ConvertNedlibOptions) Validate() error {
	if len(o.Paths) == 0 {
		return fmt.Errorf("missing file or directory name")
	}
	return nil
}

func (o *ConvertNedlibOptions) Run() error {
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

func (o *ConvertNedlibOptions) handleFile(fs afero.Fs, fileName string) (stat.Result, error) {
	result := stat.NewResult(fileName)

	nedlibReader, err := nedlibreader.NewNedlibReader(fs, fileName, o.WarcWriterConfig.DefaultTime, o.WarcRecordOptions...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = nedlibReader.Close() }()

	var writer *gowarc.WarcFileWriter
	if o.WarcWriterConfig.OneToOneWriter {
		defer func() {
			if writer != nil {
				_ = writer.Close()
			}
		}()
	}

	for record, err := range warc.Records(nedlibReader, nil, 0, 0) {
		if err != nil {
			return result, warc.Error(record, err)
		}

		warcDate, err := record.WarcRecord.WarcHeader().GetTime(gowarc.WarcDate)
		if err != nil {
			return result, warc.Error(record, err)
		}

		syntheticFileName := time.To14(warcDate)

		writer, err := o.WarcWriterConfig.GetWarcWriter(syntheticFileName, warcDate)
		if err != nil {
			return result, warc.Error(record, err)
		}

		err = o.handleRecord(writer, record, result)
		if err != nil {
			return result, warc.Error(record, err)
		}
	}

	return result, nil
}

func (o *ConvertNedlibOptions) handleRecord(w *gowarc.WarcFileWriter, record warc.Record, result stat.Result) error {
	defer record.Close()

	result.IncrRecords()

	if !record.Validation.Valid() {
		for _, err := range *record.Validation {
			result.AddError(warc.Error(record, err))
		}
	}

	writeResponse := w.Write(record.WarcRecord)
	if len(writeResponse) > 0 {
		return writeResponse[0].Err
	}
	return nil
}
