package nedlib

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/v3/cmd/internal/flag"
	"github.com/nlnwa/warchaeology/v3/cmd/internal/log"
	"github.com/nlnwa/warchaeology/v3/internal/filewalker"
	"github.com/nlnwa/warchaeology/v3/internal/hooks"
	"github.com/nlnwa/warchaeology/v3/internal/index"
	"github.com/nlnwa/warchaeology/v3/internal/stat"
	"github.com/nlnwa/warchaeology/v3/internal/time"
	"github.com/nlnwa/warchaeology/v3/internal/util"
	"github.com/nlnwa/warchaeology/v3/internal/version"
	"github.com/nlnwa/warchaeology/v3/internal/warc"
	"github.com/nlnwa/warchaeology/v3/internal/warcwriterconfig"
	"github.com/nlnwa/warchaeology/v3/internal/workerpool"
	"github.com/nlnwa/warchaeology/v3/nedlibreader"
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
	WarcRecordOptions  []gowarc.WarcRecordOption
	OpenInputFileHook  hooks.OpenInputFileHook
	CloseInputFileHook hooks.CloseInputFileHook
}

type ConvertNedlibFlags struct {
	FileWalkerFlags       flag.FileWalkerFlags
	WarcWriterConfigFlags flag.WarcWriterConfigFlags
	IndexFlags            flag.IndexFlags
	OutputHookFlags       flag.OutputHookFlags
	InputHookFlags        flag.InputHookFlags
	ConcurrencyFlags      flag.ConcurrencyFlags
}

func (f ConvertNedlibFlags) AddFlags(cmd *cobra.Command) {
	f.FileWalkerFlags.AddFlags(cmd, flag.WithDefaultSuffixes([]string{".meta"}))
	f.WarcWriterConfigFlags.AddFlags(cmd, flag.WithDefaultFilePrefix("nedlib_"), flag.WithCmdName(cmd.Name()))
	f.IndexFlags.AddFlags(cmd, flag.WithDefaultIndexSubDir(cmd.Name()))
	f.OutputHookFlags.AddFlags(cmd)
	f.InputHookFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)
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

	warcWriterConfig.WarcInfoFunc = func(wr gowarc.WarcRecordBuilder) error {
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
	}, nil
}

func NewCmdConvertNedlib() *cobra.Command {
	flags := ConvertNedlibFlags{}

	cmd := &cobra.Command{
		Use:   "nedlib <files/dirs>",
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
			return o.Run(cmd.Name(), args)
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

func (o *ConvertNedlibOptions) Run(cmd string, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if o.FileIndex != nil {
		defer o.FileIndex.Close()
	}
	defer o.WarcWriterConfig.Close()

	closer, err := log.InitLogger(os.Stderr)
	if err != nil {
		return err
	}
	defer closer.Close()

	workerPool := workerpool.New(o.Concurrency)
	defer workerPool.CloseWait()

	walkFn := func(fs afero.Fs, path string, err error) error {
		if err != nil {
			return err
		}

		slog := slog.With("path", path)

		return workerPool.Submit(ctx, func() {
			if o.MinWARCDiskFree > 0 {
				diskFree := util.DiskFree(o.WarcWriterConfig.OutDir)
				if diskFree < o.MinWARCDiskFree {
					cancel()
					slog.Error("not enough space left on device", "space", diskFree, "dir", o.WarcWriterConfig.OutDir)
					return
				}
			}

			result, err := filewalker.Preposterous(path, o.FileIndex, o.OpenInputFileHook, o.CloseInputFileHook, func() stat.Result {
				return o.readFile(ctx, fs, path)
			})
			if errors.Is(err, filewalker.ErrSkipFile) {
				return
			}
			if err != nil {
				cancel()
				slog.Error(err.Error())
				return
			}
			for _, err := range result.Errors() {
				slog.Error(err.Error())
			}
			if result.Fatal() != nil {
				cancel()
				slog.Error(result.Fatal().Error())
				return
			}
		})
	}

	for _, path := range o.Paths {
		err := o.FileWalker.Walk(path, walkFn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *ConvertNedlibOptions) readFile(ctx context.Context, fs afero.Fs, fileName string) stat.Result {
	result := stat.NewResult(fileName)

	nedlibReader, err := nedlibreader.NewNedlibReader(fs, fileName, o.WarcWriterConfig.DefaultTime, o.WarcRecordOptions...)
	if err != nil {
		result.SetFatal(err)
		return result
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

	for record := range warc.NewIterator(ctx, nedlibReader, nil, 0, 0) {
		if record.Err != nil {
			result.SetFatal(warc.Error(record, record.Err))
			break
		}

		warcDate := record.WarcRecord.WarcHeader().Get(gowarc.WarcDate)

		syntheticFileName, err := time.To14(warcDate)
		if err != nil {
			result.SetFatal(warc.Error(record, err))
			break
		}

		writer, err := o.WarcWriterConfig.GetWarcWriter(syntheticFileName, warcDate)
		if err != nil {
			result.SetFatal(warc.Error(record, err))
			break
		}

		o.handleRecord(writer, record, result)
		if result.Fatal() != nil {
			break
		}
	}

	return result
}

// handleRecord processes one record
func (o *ConvertNedlibOptions) handleRecord(w *gowarc.WarcFileWriter, record warc.Record, result stat.Result) {
	defer record.Close()

	result.IncrRecords()
	result.IncrProcessed()

	if record.Err != nil {
		result.SetFatal(warc.Error(record, record.Err))
		return
	}

	if !record.Validation.Valid() {
		result.AddError(warc.Error(record, record.Validation))
	}

	if writeResponse := w.Write(record.WarcRecord); len(writeResponse) > 0 && writeResponse[0].Err != nil {
		result.SetFatal(warc.Error(record, writeResponse[0].Err))
	}
}
