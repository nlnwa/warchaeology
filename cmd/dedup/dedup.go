package dedup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
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

const (
	BufferMaxMem     = "max-buffer-mem"
	BufferMaxMemHelp = "the maximum bytes of memory allowed for each buffer before overflowing to disk"

	DedupSizeGain     = "min-size-gain"
	DedupSizeGainHelp = `minimum bytes one must earn to perform a deduplication`

	MinIndexDiskFree     = "min-index-disk-free"
	MinIndexDiskFreeHelp = `minimum free space on disk to allow index writing`

	RecordTypes     = "record-types"
	RecordTypesHelp = `comma separated list of record types to deduplicate. Other record types are written as is.`
)

type DedupOptions struct {
	Paths               []string
	Concurrency         int
	DigestIndex         *index.DigestIndex
	FileIndex           *index.FileIndex
	WarcWriterConfig    *warcwriterconfig.WarcWriterConfig
	MinimumSizeGain     int64
	MinWARCDiskFree     int64
	MinIndexDiskFree    int64
	ContinueOnError     bool
	RecordTypes         []gowarc.RecordType
	FileWalker          *filewalker.FileWalker
	WarcRecordOptions   []gowarc.WarcRecordOption
	OpenInputFileHook   hooks.OpenInputFileHook
	CloseInputFileHook  hooks.CloseInputFileHook
	OpenOutputFileHook  hooks.OpenOutputFileHook
	CloseOutputFileHook hooks.CloseOutputFileHook
}

type DedupFlags struct {
	WarcFileFlags         flag.WarcIteratorFlags
	InputHookFlags        *flag.InputHookFlags
	OutputHookFlags       *flag.OutputHookFlags
	FileWalkerFlags       flag.FileWalkerFlags
	WarcWriterConfigFlags *flag.WarcWriterConfigFlags
	UtilFlags             flag.UtilFlags
	RepairFlags           flag.RepairFlags
	WarcRecordOptionFlags flag.WarcRecordOptionFlags
	IndexFlags            *flag.IndexFlags
	ConcurrencyFlags      flag.ConcurrencyFlags
	ErrorFlags            flag.ErrorFlags
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
	f.WarcFileFlags.AddFlags(cmd)
	f.InputHookFlags.AddFlags(cmd)
	f.OutputHookFlags.AddFlags(cmd)
	f.FileWalkerFlags.AddFlags(cmd)
	f.WarcWriterConfigFlags.AddFlags(cmd)
	f.UtilFlags.AddFlags(cmd)
	f.RepairFlags.AddFlags(cmd)
	f.WarcRecordOptionFlags.AddFlags(cmd)
	f.IndexFlags.AddFlags(cmd)
	f.ConcurrencyFlags.AddFlags(cmd)
	f.ErrorFlags.AddFlags(cmd)

	flags := cmd.Flags()
	flags.String(BufferMaxMem, "1MB", BufferMaxMemHelp)
	flags.StringP(DedupSizeGain, "g", "2KB", DedupSizeGainHelp)
	flags.String(MinIndexDiskFree, "1 * 1024 * 1024", MinIndexDiskFreeHelp)
	flags.StringSlice(RecordTypes, []string{"response", "resource"}, RecordTypesHelp)
}

func (f DedupFlags) BufferMaxMem() int64 {
	return util.ParseSizeInBytes(viper.GetString(BufferMaxMem))
}

func (f DedupFlags) DedupSizeGain() int64 {
	return util.ParseSizeInBytes(viper.GetString(DedupSizeGain))
}

func (f DedupFlags) MinIndexDiskFree() int64 {
	return util.ParseSizeInBytes(viper.GetString(MinIndexDiskFree))
}

func (f DedupFlags) RecordTypes() []string {
	return viper.GetStringSlice(RecordTypes)
}

func (f DedupFlags) ToDedupOptions() (*DedupOptions, error) {
	var recordTypes []gowarc.RecordType
	for _, rt := range f.RecordTypes() {
		recordType := stringToRecordType(rt)
		if recordType == gowarc.Revisit {
			return nil, fmt.Errorf("revisit records cannot be deduplicated")
		}
		if recordType == 0 {
			return nil, fmt.Errorf("invalid record type: %s", rt)
		}
		recordTypes = append(recordTypes, stringToRecordType(rt))
	}

	concurrency := f.ConcurrencyFlags.Concurrency()

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

	if f.RepairFlags.Repair() {
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
		Paths:              fileList,
		Concurrency:        concurrency,
		MinimumSizeGain:    f.DedupSizeGain(),
		MinWARCDiskFree:    f.UtilFlags.MinFreeDisk(),
		MinIndexDiskFree:   f.MinIndexDiskFree(),
		FileWalker:         fileWalker,
		ContinueOnError:    f.ErrorFlags.ContinueOnError(),
		RecordTypes:        recordTypes,
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
		Use:   "dedup FILE/DIR ...",
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

func (o *DedupOptions) Complete(cmd *cobra.Command, args []string) error {
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
	if len(o.RecordTypes) == 0 {
		return errors.New("missing record type(s)")
	}
	return nil
}

func (o *DedupOptions) Run() error {
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
			slog.Info("Total", "errors", stats.Errors, "records", stats.Records, "duplicates", stats.Duplicates)
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
			slog.Info("Deduplicated file", "errors", result.ErrorCount(), "records", result.Records(), "duplicates", result.Duplicates())
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
	if o.DigestIndex != nil {
		defer o.DigestIndex.Close()
	}

	defer o.WarcWriterConfig.Close()

	workerPool := workerpool.New(ctx, o.Concurrency)
	defer workerPool.CloseWait()

	// Warn when file descriptor limit is below recommended value
	err := util.CheckFileDescriptorLimit(util.BadgerRecommendedMaxFileDescr)
	if err != nil {
		slog.Warn(err.Error())
	}

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
				// Assert index disk has enough free space
				if o.MinIndexDiskFree > 0 {
					diskFree, err := util.DiskFree(o.DigestIndex.GetDir())
					if err != nil {
						cancel()
						slog.Error("Failed to get free space on device", "path", o.DigestIndex.GetDir(), "error", err)
						return
					}
					if diskFree < o.MinIndexDiskFree {
						cancel()
						slog.Error("Not enough free space on device", "bytesFree", diskFree, "path", o.DigestIndex.GetDir())
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

func (o *DedupOptions) handleFile(fs afero.Fs, path string) (stat.Result, error) {
	result := stat.NewResult(path)

	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	warcReader, err := gowarc.NewWarcFileReaderFromStream(file, 0, o.WarcRecordOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create warc file reader: %w", err)
	}
	defer func() { _ = warcReader.Close() }()

	var writer *gowarc.WarcFileWriter
	if o.WarcWriterConfig.OneToOneWriter {
		defer func() {
			if writer != nil {
				_ = writer.Close()
			}
		}()
	}

	for record, err := range warc.Records(warcReader, nil, 0, 0) {
		if err != nil {
			return result, warc.Error(record, err)
		}

		if writer == nil || !o.WarcWriterConfig.OneToOneWriter {
			warcDate := record.WarcRecord.WarcHeader().Get(gowarc.WarcDate)
			writer, err = o.WarcWriterConfig.GetWarcWriter(path, warcDate)
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

func (o *DedupOptions) handleRecord(writer *gowarc.WarcFileWriter, record warc.Record, result stat.Result) error {
	defer record.Close()

	result.IncrRecords()

	warcRecord := record.WarcRecord

	// if the record type is not in the list of record types to deduplicate, write the record as is
	if !slices.Contains(o.RecordTypes, warcRecord.Type()) {
		return writeRecord(writer, warcRecord)
	}

	digest, err := getDigest(warcRecord, record.Validation)
	if err != nil {
		result.AddError(fmt.Errorf("failed to get digest: %w", err))
	}
	if !record.Validation.Valid() {
		result.AddError(record.Validation)
	}

	// write response if no digest is found because we can't make a revisit record without a digest)
	if digest == "" {
		return writeRecord(writer, warcRecord)
	}

	profile := getRevisitProfile(warcRecord)

	// Write response if no profile is found (can't make a revisit record without a profile)
	if profile == "" {
		return writeRecord(writer, warcRecord)
	}

	revisitRef, err := warcRecord.CreateRevisitRef(profile)
	if err != nil {
		result.AddError(fmt.Errorf("error creating revisit ref: %w", err))
	}

	// determine if the record is a revisit record
	revisitReference, err := o.DigestIndex.IsRevisit(digest, revisitRef)
	if err != nil {
		result.AddError(fmt.Errorf("error getting revisit ref: %w", err))
	}

	// write a normal record if one of:
	// a - no revisit reference is found
	// b - the revisit reference is too large compared to the original record (not enough size gain to warrant writing a revisit record)
	if revisitReference == nil ||
		int64(revisitRefSize(revisitReference)) >= payloadLength(warcRecord)-o.MinimumSizeGain {
		return writeRecord(writer, warcRecord)
	}

	// Make a revisit record. If it fails, write the original record.
	revisit, err := warcRecord.ToRevisitRecord(revisitReference)
	if err != nil {
		result.AddError(fmt.Errorf("error creating revisit record: %w", err))
		return writeRecord(writer, warcRecord)
	}

	// Write revisit record
	err = writeRecord(writer, revisit)
	if err == nil {
		result.IncrDuplicates()
	}
	return err
}

func writeRecord(writer *gowarc.WarcFileWriter, warcRecord gowarc.WarcRecord) error {
	writeResponse := writer.Write(warcRecord)
	if len(writeResponse) > 0 {
		return writeResponse[0].Err
	}
	return nil
}

// getRevisitProfile returns the revisit profile for a record
func getRevisitProfile(warcRecord gowarc.WarcRecord) string {
	switch warcRecord.Version() {
	case gowarc.V1_0:
		return gowarc.ProfileIdenticalPayloadDigestV1_0
	case gowarc.V1_1:
		return gowarc.ProfileIdenticalPayloadDigestV1_1
	default:
		return ""
	}
}

// getDigest returns the digest of a record. If the record does not have a digest, it will be calculated and validated.
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

func stringToRecordType(rt string) gowarc.RecordType {
	switch rt {
	case "warcinfo":
		return gowarc.Warcinfo
	case "response":
		return gowarc.Response
	case "resource":
		return gowarc.Resource
	case "request":
		return gowarc.Request
	case "metadata":
		return gowarc.Metadata
	case "revisit":
		return gowarc.Revisit
	case "conversion":
		return gowarc.Conversion
	case "continuation":
		return gowarc.Continuation
	default:
		return 0
	}
}
