package dedup

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
	"github.com/nlnwa/warchaeology/v3/internal/filter"
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

const (
	BufferMaxMem     = "max-buffer-mem"
	BufferMaxMemHelp = "the maximum bytes of memory allowed for each buffer before overflowing to disk"

	DedupSizeGain     = "min-size-gain"
	DedupSizeGainHelp = `minimum bytes one must earn to perform a deduplication`
)

type DedupOptions struct {
	Filter              *filter.RecordFilter
	Paths               []string
	Concurrency         int
	DigestIndex         *index.DigestIndex
	FileIndex           *index.FileIndex
	WarcWriterConfig    *warcwriterconfig.WarcWriterConfig
	MinimumSizeGain     int64
	MinWARCDiskFree     int64
	Repair              bool
	FileWalker          *filewalker.FileWalker
	WarcRecordOptions   []gowarc.WarcRecordOption
	OpenInputFileHook   hooks.OpenInputFileHook
	CloseInputFileHook  hooks.CloseInputFileHook
	OpenOutputFileHook  hooks.OpenOutputFileHook
	CloseOutputFileHook hooks.CloseOutputFileHook
}

type DedupFlags struct {
	FilterFlags           flag.FilterFlags
	WarcFileFlags         flag.WarcIteratorFlags
	InputHookFlags        *flag.InputHookFlags
	OutputHookFlags       *flag.OutputHookFlags
	FileWalkerFlags       flag.FileWalkerFlags
	WarcWriterConfigFlags *flag.WarcWriterConfigFlags
	UtilFlags             flag.UtilFlags
	WarcRecordOptionFlags flag.WarcRecordOptionFlags
	IndexFlags            *flag.IndexFlags
	ConcurrencyFlags      flag.ConcurrencyFlags
}

func NewDedupFlags() DedupFlags {
	return DedupFlags{
		InputHookFlags:        &flag.InputHookFlags{},
		OutputHookFlags:       &flag.OutputHookFlags{},
		WarcWriterConfigFlags: &flag.WarcWriterConfigFlags{},
		IndexFlags:            &flag.IndexFlags{},
	}
}

func (f DedupFlags) AddFlags(cmd *cobra.Command) {
	f.FilterFlags.AddFlags(cmd)
	f.WarcFileFlags.AddFlags(cmd)
	f.InputHookFlags.AddFlags(cmd)
	f.OutputHookFlags.AddFlags(cmd)
	f.FileWalkerFlags.AddFlags(cmd)
	f.WarcWriterConfigFlags.AddFlags(cmd)
	f.UtilFlags.AddFlags(cmd)
	f.WarcRecordOptionFlags.AddFlags(cmd)
	f.IndexFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)

	flags := cmd.Flags()
	flags.String(BufferMaxMem, "1MB", BufferMaxMemHelp)
	flags.StringP(DedupSizeGain, "g", "2KB", DedupSizeGainHelp)
}

func (f DedupFlags) BufferMaxMem() int64 {
	return util.ParseSizeInBytes(viper.GetString(BufferMaxMem))
}

func (f DedupFlags) DedupSizeGain() int64 {
	return util.ParseSizeInBytes(viper.GetString(DedupSizeGain))
}

func (f DedupFlags) ToDedupOptions() (*DedupOptions, error) {
	concurrency := f.ConcurrencyFlags.Concurrency()

	filter, err := f.FilterFlags.ToFilter()
	if err != nil {
		return nil, err
	}

	warcWriterConfig, err := f.WarcWriterConfigFlags.ToWarcWriterConfig()
	if err != nil {
		return nil, err
	}

	warcRecordOptions := []gowarc.WarcRecordOption{
		gowarc.WithBufferTmpDir(f.WarcRecordOptionFlags.TmpDir()),
		gowarc.WithBufferMaxMemBytes(f.BufferMaxMem()),
		gowarc.WithSyntaxErrorPolicy(gowarc.ErrWarn),
		gowarc.WithSpecViolationPolicy(gowarc.ErrWarn),
		gowarc.WithAddMissingDigest(true),
		gowarc.WithAddMissingContentLength(true),
	}

	if f.UtilFlags.Repair() {
		warcRecordOptions = append(warcRecordOptions,
			gowarc.WithFixSyntaxErrors(true),
			gowarc.WithFixDigest(true),
			gowarc.WithAddMissingRecordId(true),
			gowarc.WithFixContentLength(true),
			gowarc.WithFixWarcFieldsBlockErrors(true),
		)
	} else {
		warcRecordOptions = append(warcRecordOptions,
			gowarc.WithFixSyntaxErrors(false),
			gowarc.WithFixDigest(false),
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

	digestIndex, err := f.IndexFlags.ToDigestIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to create digest index: %w", err)
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

	return &DedupOptions{
		Filter:             filter,
		Paths:              fileList,
		Concurrency:        concurrency,
		MinimumSizeGain:    f.DedupSizeGain(),
		MinWARCDiskFree:    f.UtilFlags.MinFreeDisk(),
		Repair:             f.UtilFlags.Repair(),
		FileWalker:         fileWalker,
		WarcWriterConfig:   warcWriterConfig,
		WarcRecordOptions:  warcRecordOptions,
		DigestIndex:        digestIndex,
		FileIndex:          fileIndex,
		OpenInputFileHook:  openInputFileHook,
		CloseInputFileHook: closeInputFileHook,
	}, nil
}

func NewCmdDedup() *cobra.Command {
	flags := NewDedupFlags()

	var cmd = &cobra.Command{
		Use:   "dedup",
		Short: "Deduplicate WARC files",
		Long: `Deduplicate WARC files.

NOTE: The filtering options only decides which records are candidates for deduplication.
The remaining records are written as is.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToDedupOptions()
			if err != nil {
				return err
			}
			if err := o.Complete(cmd, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
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

func (o *DedupOptions) Complete(cmd *cobra.Command, args []string) error {
	err := util.CheckFileDescriptorLimit(util.BadgerRecommendedMaxFileDescr)
	if err != nil {
		slog.Warn(err.Error())
	}

	o.Paths = append(o.Paths, args...)

	return nil
}

func (o *DedupOptions) Validate() error {
	if o.WarcWriterConfig.OutDir == "" {
		return errors.New("missing output directory")
	}
	if len(o.Paths) == 0 {
		return errors.New("missing file or directory name")
	}

	return nil
}

func (o *DedupOptions) Run() error {
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
	if o.DigestIndex != nil {
		defer o.DigestIndex.Close()
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

			if err = o.DigestIndex.HasDiskSpace(); err != nil {
				slog.Error(err.Error())
				return
			}

			result, err := filewalker.Preposterous(path, o.FileIndex, o.OpenInputFileHook, o.CloseInputFileHook, func() stat.Result {
				return o.dedupFile(ctx, fs, path)
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

func (o *DedupOptions) dedupFile(ctx context.Context, fs afero.Fs, fileName string) stat.Result {
	result := stat.NewResult(fileName)

	file, err := fs.Open(fileName)
	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = file.Close() }()

	warcReader, err := gowarc.NewWarcFileReaderFromStream(file, 0, o.WarcRecordOptions...)
	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = warcReader.Close() }()

	var warcWriter *gowarc.WarcFileWriter
	if o.WarcWriterConfig.OneToOneWriter {
		defer func() {
			if warcWriter != nil {
				_ = warcWriter.Close()
			}
		}()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for record := range warc.NewIterator(ctx, warcReader, o.Filter, 0, 0) {
		if warcWriter == nil {
			warcDate := record.WarcRecord.WarcHeader().Get(gowarc.WarcDate)
			warcWriter, err = o.WarcWriterConfig.GetWarcWriter(fileName, warcDate)
			if err != nil {
				result.SetFatal(warc.Error(record, err))
				break
			}
		}
		o.handleRecord(warcWriter, record, result)
		if result.Fatal() != nil {
			break
		}
		if !o.WarcWriterConfig.OneToOneWriter {
			warcWriter = nil
		}
	}

	return result
}

// handleRecord processes one record
// The input parameter writer and output parameter writerOut is only used for identity transformation. In this case there is one writer
// per file which should be closed by readFile when the file is processed. But since we need the warc date of the first record to open
// the writer, it must be opened in this function. These parameters are used for giving readFile access to the writer.
func (o *DedupOptions) handleRecord(writer *gowarc.WarcFileWriter, record warc.Record, result stat.Result) {
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

	warcRecord := record.WarcRecord

	length := payloadLength(warcRecord)
	digest, err := getDigest(warcRecord, record.Validation)
	if err != nil {
		result.AddError(warc.Error(record, fmt.Errorf("failed to get digest: %w", err)))
	}
	if digest == "" {
		err = writeRecord(writer, warcRecord)
		if err != nil {
			result.SetFatal(warc.Error(record, fmt.Errorf("failed to write record: %w", err)))
			return
		}
	}

	profile := getRevisitProfile(warcRecord)
	revisitRef, _ := warcRecord.CreateRevisitRef(profile)

	revisitReference, err := o.DigestIndex.IsRevisit(digest, revisitRef)
	if err != nil {
		result.AddError(warc.Error(record, fmt.Errorf("error getting revisit ref: %w", err)))
	}
	// write a normal record if one of:
	// a - no revisit reference is found
	// b - the revisit reference is too large compared to the original record (not enough size gain to warrant writing a revisit record)
	if revisitReference == nil ||
		int64(revisitRefSize(revisitReference)) >= length-o.MinimumSizeGain {

		err = writeRecord(writer, warcRecord)
		if err != nil {
			result.SetFatal(warc.Error(record, fmt.Errorf("failed to write record: %w", err)))
		}

		return
	}

	if revisitReference.Profile == "" {
		result.AddError(warc.Error(record, fmt.Errorf("revisit reference has no profile: %v", revisitReference)))
	}

	// Write a revisit record. If it fails, write the original record.
	revisit, err := warcRecord.ToRevisitRecord(revisitReference)
	if err != nil {
		result.AddError(warc.Error(record, fmt.Errorf("error creating revisit record: %w", err)))
		err = writeRecord(writer, warcRecord)
		if err != nil {
			result.SetFatal(warc.Error(record, fmt.Errorf("failed to write record: %w", err)))
		}
		return
	}
	err = writeRecord(writer, revisit)
	if err != nil {
		result.SetFatal(warc.Error(record, fmt.Errorf("failed to write record: %w", err)))
	} else {
		result.IncrDuplicates()
	}
}

func writeRecord(writer *gowarc.WarcFileWriter, warcRecord gowarc.WarcRecord) error {
	writeResponse := writer.Write(warcRecord)
	if len(writeResponse) == 0 {
		return nil
	}
	if writeResponse[0].Err != nil {
		return writeResponse[0].Err
	}
	return nil
}

func getRevisitProfile(warcRecord gowarc.WarcRecord) string {
	switch warcRecord.Version() {
	case gowarc.V1_0:
		return gowarc.ProfileIdenticalPayloadDigestV1_0
	default:
		return gowarc.ProfileIdenticalPayloadDigestV1_1
	}
}

func getDigest(warcRecord gowarc.WarcRecord, validation *gowarc.Validation) (string, error) {
	var digest string
	if warcRecord.WarcHeader().Has(gowarc.WarcPayloadDigest) {
		digest = warcRecord.WarcHeader().Get(gowarc.WarcPayloadDigest)
	} else if warcRecord.WarcHeader().Has(gowarc.WarcBlockDigest) {
		digest = warcRecord.WarcHeader().Get(gowarc.WarcBlockDigest)
	} else {
		if err := warcRecord.Block().Cache(); err != nil {
			return digest, fmt.Errorf("could not cache record: %w", err)
		}
		if err := warcRecord.ValidateDigest(validation); err != nil {
			return digest, fmt.Errorf("failed to validate digest: %w", err)
		}
		return getDigest(warcRecord, validation)
	}
	return digest, nil
}

func payloadLength(warcRecord gowarc.WarcRecord) int64 {
	var length int64
	switch block := warcRecord.Block().(type) {
	case gowarc.ProtocolHeaderBlock:
		length, _ = warcRecord.ContentLength()
		length -= int64(len(block.ProtocolHeaderBytes()))
	case gowarc.WarcFieldsBlock:
		length = block.Size()
	}
	return length
}

func revisitRefSize(revisitReference *gowarc.RevisitRef) int {
	length := 0
	if revisitReference.TargetRecordId != "" {
		length += len(gowarc.WarcRefersTo) + len(revisitReference.TargetRecordId)
	}
	if revisitReference.Profile != "" {
		length += len(gowarc.WarcProfile) + len(revisitReference.Profile)
	}
	if revisitReference.TargetUri != "" {
		length += len(gowarc.WarcRefersToTargetURI) + len(revisitReference.TargetUri)
	}
	if revisitReference.TargetDate != "" {
		length += len(gowarc.WarcRefersToDate) + len(revisitReference.TargetDate)
	}
	return length
}
