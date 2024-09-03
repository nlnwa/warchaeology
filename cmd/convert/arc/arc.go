package arc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/arcreader"
	"github.com/nlnwa/warchaeology/cmd/internal/flag"
	"github.com/nlnwa/warchaeology/cmd/internal/log"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/hooks"
	"github.com/nlnwa/warchaeology/internal/index"
	"github.com/nlnwa/warchaeology/internal/stat"
	"github.com/nlnwa/warchaeology/internal/util"
	"github.com/nlnwa/warchaeology/internal/version"
	"github.com/nlnwa/warchaeology/internal/warc"
	"github.com/nlnwa/warchaeology/internal/warcwriterconfig"
	"github.com/nlnwa/warchaeology/internal/workerpool"
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
	FileIndex          *index.FileIndex
	OpenInputFileHook  hooks.OpenInputFileHook
	CloseInputFileHook hooks.CloseInputFileHook
}

type ConvertArcFlags struct {
	FileWalkerFlags       flag.FileWalkerFlags
	IndexFlags            flag.IndexFlags
	WarcWriterConfigFlags flag.WarcWriterConfigFlags
	WarcRecordOptionFlags flag.WarcRecordOptionFlags
	OutputHookFlags       flag.OutputHookFlags
	InputHookFlags        flag.InputHookFlags
	UtilFlags             flag.UtilFlags
	ConcurrencyFlags      flag.ConcurrencyFlags
}

func (f ConvertArcFlags) AddFlags(cmd *cobra.Command) {
	f.FileWalkerFlags.AddFlags(cmd, flag.WithDefaultSuffixes([]string{".arc", ".arc.gz"}))
	f.OutputHookFlags.AddFlags(cmd)
	f.InputHookFlags.AddFlags(cmd)
	f.WarcWriterConfigFlags.AddFlags(cmd, flag.WithCmdName(cmd.Name()))
	f.WarcRecordOptionFlags.AddFlags(cmd)
	f.IndexFlags.AddFlags(cmd, flag.WithDefaultIndexSubDir(cmd.Name()))
	f.UtilFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)
}

func (f ConvertArcFlags) ToConvertArcOptions() (*ConvertArcOptions, error) {
	warcWriterConfig, err := f.WarcWriterConfigFlags.ToWarcWriterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create warc writer config: %w", err)
	}
	if warcWriterConfig.OneToOneWriter {
		warcWriterConfig.WarcInfoFunc = func(recordBuilder gowarc.WarcRecordBuilder) error {
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
	}

	warcRecordOptions := f.WarcRecordOptionFlags.ToWarcRecordOptions()
	warcRecordOptions = append(warcRecordOptions,
		gowarc.WithVersion(warcWriterConfig.WarcVersion),
		gowarc.WithAddMissingDigest(true),
	)

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
		Concurrency:       f.ConcurrencyFlags.Concurrency(),
		WarcWriterConfig:  warcWriterConfig,
		WarcRecordOptions: warcRecordOptions,
		FileWalker:        fileWalker,
		FileIndex:         fileIndex,
		Paths:             fileList,
	}, nil
}

func NewCommand() *cobra.Command {
	flags := ConvertArcFlags{}

	var cmd = &cobra.Command{
		Use:   "arc <files/dirs>",
		Short: "Convert arc file into warc file",
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
			return o.Run()
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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	closer, err := log.InitLogger(os.Stderr)
	if err != nil {
		return err
	}
	defer closer.Close()

	if o.FileIndex != nil {
		defer o.FileIndex.Close()
	}
	defer o.WarcWriterConfig.Close()

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
					slog.Error("not enough space left on device", "space", diskFree, "dir", o.WarcWriterConfig.OutDir)
					cancel()
				}
			}

			result, err := filewalker.Preposterous(path, o.FileIndex, o.OpenInputFileHook, o.CloseInputFileHook, func() stat.Result {
				return o.readFile(ctx, fs, path)
			})
			if errors.Is(err, filewalker.ErrSkipFile) {
				return
			}
			if err != nil {
				defer cancel()
				slog.Error(err.Error(), "path", path)
				return
			}
			for _, err := range result.Errors() {
				slog.Error(err.Error(), "path", path)
			}
			if result.Fatal() != nil {
				defer cancel()
				slog.Error(result.Fatal().Error(), "path", path)
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

func (o *ConvertArcOptions) readFile(ctx context.Context, fileSystem afero.Fs, fileName string) stat.Result {
	result := stat.NewResult(fileName)

	arcFileReader, err := arcreader.NewArcFileReader(fileSystem, fileName, 0, o.WarcRecordOptions...)
	if err != nil {
		result.AddError(err)
		return result
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

	for record := range warc.NewIterator(ctx, arcFileReader, nil, 0, 0) {
		if writer == nil {
			warcDate := record.WarcRecord.WarcHeader().Get(gowarc.WarcDate)
			writer, err = o.WarcWriterConfig.GetWarcWriter(fileName, warcDate)
			if err != nil {
				result.SetFatal(warc.Error(record, err))
				break
			}
		}
		o.handleRecord(writer, record, result)
		if result.Fatal() != nil {
			break
		}
		if !o.WarcWriterConfig.OneToOneWriter {
			writer = nil
		}
	}

	return result
}

// handleRecord processes one record
// The input parameter writer and output parameter writerOut is only used for identity transformation. In this case there is one writer
// per file which should be closed by readFile when the file is processed. But since we need the warc date of the first record to open
// the writer, it must be opened in this function. These parameters are used for giving readFile access to the writer.
func (o *ConvertArcOptions) handleRecord(warcFileWriter *gowarc.WarcFileWriter, record warc.Record, result stat.Result) {
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
	if writeResponse := warcFileWriter.Write(record.WarcRecord); len(writeResponse) > 0 && writeResponse[0].Err != nil {
		result.SetFatal(warc.Error(record, writeResponse[0].Err))
	}
}
