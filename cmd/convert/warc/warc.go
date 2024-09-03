package warc

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
	WarcWriterConfigFlags flag.WarcWriterConfigFlags
	OutputHookFlags       flag.OutputHookFlags
	InputHookFlags        flag.InputHookFlags
	UtilFlags             flag.UtilFlags
	ConcurrencyFlags      flag.ConcurrencyFlags
}

func (f ConvertWarcFlags) AddFlags(cmd *cobra.Command) {
	f.FileWalkerFlags.AddFlags(cmd, flag.WithDefaultSuffixes([]string{".warc", ".warc.gz"}))
	f.IndexFlags.AddFlags(cmd, flag.WithDefaultIndexSubDir(cmd.Name()))
	f.WarcRecordOptionFlags.AddFlags(cmd)
	f.WarcWriterConfigFlags.AddFlags(cmd, flag.WithCmdName(cmd.Name()))
	f.OutputHookFlags.AddFlags(cmd)
	f.InputHookFlags.AddFlags(cmd)
	f.UtilFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)
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
	if f.UtilFlags.Repair() {
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
		OpenInputFileHook:  openInputFileHook,
		CloseInputFileHook: closeInputFileHook,
	}, nil
}

func NewCommand() *cobra.Command {
	flags := ConvertWarcFlags{}

	var cmd = &cobra.Command{
		Use:   "warc <files/dirs>",
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
			return o.Run()
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
				defer cancel()
				slog.Error(err.Error())
				return
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

	for _, path := range o.Paths {
		err := o.FileWalker.Walk(path, walkFn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *ConvertWarcOptions) readFile(ctx context.Context, fileSystem afero.Fs, fileName string) stat.Result {
	result := stat.NewResult(fileName)

	file, err := fileSystem.Open(fileName)
	if err != nil {
		result.SetFatal(err)
		return result
	}
	defer func() { _ = file.Close() }()

	warcFileReader, err := gowarc.NewWarcFileReaderFromStream(file, 0, o.WarcRecordOptions...)
	if err != nil {
		result.AddError(err)
		return result
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

	for record := range warc.NewIterator(ctx, warcFileReader, nil, 0, 0) {
		if writer == nil {
			writer, err = o.WarcWriterConfig.GetWarcWriter(fileName, record.WarcRecord.WarcHeader().Get(gowarc.WarcDate))
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
func (o *ConvertWarcOptions) handleRecord(warcFileWriter *gowarc.WarcFileWriter, record warc.Record, result stat.Result) {
	defer record.Close()

	result.IncrRecords()
	result.IncrProcessed()

	if !record.Validation.Valid() {
		result.AddError(warc.Error(record, record.Validation))
	}
	if record.Err != nil {
		result.SetFatal(warc.Error(record, record.Err))
		return
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
		result.SetFatal(warc.Error(record, err))
		return
	}
	_, err = warcRecordBuilder.ReadFrom(ioReader)
	if err != nil {
		result.SetFatal(warc.Error(record, err))
		return
	}
	warcRecord, _, err = warcRecordBuilder.Build()
	if err != nil {
		result.SetFatal(warc.Error(record, err))
		return
	}
	if writeResponse := warcFileWriter.Write(warcRecord); len(writeResponse) > 0 && writeResponse[0].Err != nil {
		result.SetFatal(warc.Error(record, writeResponse[0].Err))
		return
	}
}
