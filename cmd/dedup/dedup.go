package dedup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/filter"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/utils"
	"github.com/nlnwa/warchaeology/internal/warcwriterconfig"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type conf struct {
	filter          *filter.Filter
	files           []string
	concurrency     int
	digestIndex     *DigestIndex
	writerConf      *warcwriterconfig.WarcWriterConfig
	minimumSizeGain int64
	minWARCDiskFree int64
	repair          bool
}

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "dedup",
		Short: "Deduplicate WARC files",
		Long: `Deduplicate WARC files.

NOTE: The filtering options only decides which records are candidates for deduplication.
The remaining records are written as is.`,
		RunE:              parseArgumentsAndCallDeduplication,
		ValidArgsFunction: flag.SuffixCompletionFn,
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	cmd.Flags().StringArray(flag.RecordId, []string{}, flag.RecordIdHelp)
	cmd.Flags().StringSliceP(flag.RecordType, "t", []string{"response"}, flag.RecordTypeHelp)
	cmd.Flags().StringP(flag.ResponseCode, "e", "", flag.ResponseCodeHelp)
	cmd.Flags().StringSliceP(flag.MimeType, "m", []string{}, flag.MimeTypeHelp)
	cmd.Flags().BoolP(flag.KeepIndex, "k", false, flag.KeepIndexHelp)
	cmd.Flags().BoolP(flag.NewIndex, "K", false, flag.NewIndexHelp)
	cmd.Flags().StringP(flag.IndexDir, "i", cacheDir+"/warc", flag.IndexDirHelp)
	cmd.Flags().BoolP(flag.Recursive, "r", false, flag.RecursiveHelp)
	cmd.Flags().BoolP(flag.FollowSymlinks, "s", false, flag.FollowSymlinksHelp)
	cmd.Flags().StringSlice(flag.Suffixes, []string{".warc", ".warc.gz"}, flag.SuffixesHelp)
	cmd.Flags().IntP(flag.Concurrency, "c", int(float32(runtime.NumCPU())*float32(1.5)), flag.ConcurrencyHelp)
	cmd.Flags().IntP(flag.ConcurrentWriters, "C", 16, flag.ConcurrentWritersHelp)
	cmd.Flags().StringP(flag.FileSize, "S", "1GB", flag.FileSizeHelp)
	cmd.Flags().BoolP(flag.Compress, "z", false, flag.CompressHelp)
	cmd.Flags().Bool(flag.CompressionLevel, false, flag.CompressionLevelHelp)
	cmd.Flags().StringP(flag.FilePrefix, "p", "", flag.FilePrefixHelp)
	cmd.Flags().StringP(flag.WarcDir, "w", ".", flag.WarcDirHelp)
	cmd.Flags().String(flag.SubdirPattern, "", flag.SubdirPatternHelp)
	cmd.Flags().StringP(flag.NameGenerator, "n", "default", flag.NameGeneratorHelp)
	cmd.Flags().Bool(flag.Flush, false, flag.FlushHelp)
	cmd.Flags().StringP(flag.DedupSizeGain, "g", "2KB", flag.DedupSizeGainHelp)
	cmd.Flags().String(flag.MinFreeDisk, "256MB", flag.MinFreeDiskHelp)
	cmd.Flags().BoolP(flag.Repair, "R", false, flag.RepairHelp)
	cmd.Flags().String(flag.SrcFilesystem, "", flag.SrcFilesystemHelp)
	cmd.Flags().String(flag.OpenInputFileHook, "", flag.OpenInputFileHookHelp)
	cmd.Flags().String(flag.CloseInputFileHook, "", flag.CloseInputFileHookHelp)
	cmd.Flags().String(flag.OpenOutputFileHook, "", flag.OpenOutputFileHookHelp)
	cmd.Flags().String(flag.CloseOutputFileHook, "", flag.CloseOutputFileHookHelp)

	if err := cmd.RegisterFlagCompletionFunc(flag.RecordType, flag.SliceCompletion{
		"warcinfo",
		"request",
		"response",
		"metadata",
		"revisit",
		"resource",
		"continuation",
		"conversion",
	}.CompletionFn); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagDirname(flag.IndexDir); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(flag.FilePrefix, cobra.NoFileCompletions); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagDirname(flag.WarcDir); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(flag.SubdirPattern, cobra.NoFileCompletions); err != nil {
		panic(err)
	}
	if err := cmd.RegisterFlagCompletionFunc(flag.NameGenerator, cobra.NoFileCompletions); err != nil {
		panic(err)
	}

	return cmd
}

func parseArgumentsAndCallDeduplication(cmd *cobra.Command, args []string) error {
	config := &conf{}
	utils.CheckFileDescriptorLimit(utils.BadgerRecommendedMaxFileDescr)

	if wc, err := warcwriterconfig.NewFromViper(cmd.Name()); err != nil {
		return err
	} else {
		config.writerConf = wc
	}

	config.concurrency = viper.GetInt(flag.Concurrency)
	config.minimumSizeGain = utils.ParseSizeInBytes(viper.GetString(flag.DedupSizeGain))
	config.minWARCDiskFree = utils.ParseSizeInBytes(viper.GetString(flag.MinFreeDisk))
	config.repair = viper.GetBool(flag.Repair)

	if len(args) == 0 && viper.GetString(flag.SrcFileList) == "" {
		return errors.New("missing file or directory name")
	}
	config.files = args
	var err error

	if config.digestIndex, err = NewDigestIndex(viper.GetBool(flag.NewIndex), cmd.Name()); err != nil {
		fmt.Println(err)
		return nil
	}
	defer config.digestIndex.Close()

	config.filter = filter.NewFromViper()

	return runE(cmd.Name(), config)
}

func runE(cmd string, c *conf) error {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	defer c.writerConf.Close()

	fw, err := filewalker.NewFromViper(cmd, c.files, c.readFile)
	if err != nil {
		return err
	}

	stats := filewalker.NewStats()
	return fw.Walk(ctx, stats)
}

func (c *conf) readFile(fs afero.Fs, fileName string) filewalker.Result {
	result := filewalker.NewResult(fileName)

	opts := []gowarc.WarcRecordOption{
		gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)),
		gowarc.WithBufferMaxMemBytes(utils.ParseSizeInBytes(viper.GetString(flag.BufferMaxMem))),
	}
	if c.repair {
		opts = append(opts,
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
		opts = append(opts,
			gowarc.WithSyntaxErrorPolicy(gowarc.ErrWarn),
			gowarc.WithSpecViolationPolicy(gowarc.ErrWarn),
			gowarc.WithAddMissingDigest(true),
			gowarc.WithFixSyntaxErrors(false),
			gowarc.WithFixDigest(false),
			gowarc.WithAddMissingContentLength(true),
			gowarc.WithAddMissingRecordId(false),
			gowarc.WithFixContentLength(false),
		)
	}

	f, err := fs.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()
	wf, err := gowarc.NewWarcFileReaderFromStream(f, 0, opts...)

	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = wf.Close() }()

	var writer *gowarc.WarcFileWriter
	if c.writerConf.WarcFileNameGenerator == "identity" {
		defer func() {
			if writer != nil {
				_ = writer.Close()
			}
		}()
	}

	for {
		if utils.DiskFree(c.writerConf.OutDir) < c.minWARCDiskFree {
			result.SetFatal(utils.NewOutOfSpaceError("cannot write WARC file, almost no space left in directory '%s'\n", c.writerConf.OutDir))
			break
		}
		var currentOffset int64
		currentOffset, writer, err = handleRecord(c, wf, fileName, result, writer)
		if err == io.EOF {
			break
		}
		if result.Fatal() != nil {
			break
		}
		if err != nil {
			result.AddError(fmt.Errorf("error: %v, rec num: %d, offset %d", err.Error(), result.Records(), currentOffset))
			break
		}
	}

	return result
}

// handleRecord processes one record
// The input parameter writer and output parameter writerOut is only used for identity transformation. In this case there is one writer
// per file which should be closed by readFile when the file is processed. But since we need the warc date of the first record to open
// the writer, it must be opened in this function. These parameters are used for giving readFile access to the writer.
func handleRecord(c *conf, wf *gowarc.WarcFileReader, fileName string, result filewalker.Result, writer *gowarc.WarcFileWriter) (offset int64, writerOut *gowarc.WarcFileWriter, err error) {
	writerOut = writer
	wr, currentOffset, validation, err := wf.Next()
	if wr != nil {
		defer func() {
			_ = wr.Close()
		}()
	}

	offset = currentOffset
	if err != nil {
		return
	}

	result.IncrRecords()
	if !validation.Valid() {
		result.AddError(fmt.Errorf("info: found problem rec num: %d, offset %d: %w", result.Records(), currentOffset, validation))
	}

	if writer == nil {
		writer = c.writerConf.GetWarcWriter(fileName, wr.WarcHeader().Get(gowarc.WarcDate))
		if c.writerConf.WarcFileNameGenerator == "identity" {
			writerOut = writer
		}
	}

	if c.filter.Accept(wr) {
		result.IncrProcessed()

		length := payloadLength(wr)
		digest := getDigest(wr, validation, result)
		profile := getRevisitProfile(wr)
		revisitRef, _ := wr.CreateRevisitRef(profile)

		r, err := c.digestIndex.IsRevisit(digest, revisitRef)
		if err != nil {
			if _, ok := err.(utils.OutOfSpaceError); ok {
				result.SetFatal(err)
			} else {
				result.AddError(fmt.Errorf("error getting revisit ref: %w", err))
			}
		}
		if r != nil && int64(revisitRefSize(r)) < length-c.minimumSizeGain {
			if r.Profile == "" {
				panic(r)
			}

			if revisit, err := wr.ToRevisitRecord(r); err != nil {
				result.AddError(fmt.Errorf("error creating revisit record: %w", err))
				writeRecord(writer, currentOffset, wr)
			} else {
				writeRecord(writer, currentOffset, revisit)
				result.IncrDuplicates()
			}
		} else {
			writeRecord(writer, currentOffset, wr)
		}
	} else {
		writeRecord(writer, currentOffset, wr)
	}
	return
}

func writeRecord(writer *gowarc.WarcFileWriter, offset int64, wr gowarc.WarcRecord) {
	if rr := writer.Write(wr); rr != nil && rr[0].Err != nil {
		fmt.Printf("Offset: %d\n", offset)
		_, _ = wr.WarcHeader().Write(os.Stdout)
		panic(rr[0].Err)
	}
}

func getRevisitProfile(wr gowarc.WarcRecord) string {
	switch wr.Version() {
	case gowarc.V1_0:
		return gowarc.ProfileIdenticalPayloadDigestV1_0
	default:
		return gowarc.ProfileIdenticalPayloadDigestV1_1
	}
}

func getDigest(wr gowarc.WarcRecord, validation *gowarc.Validation, result filewalker.Result) string {
	var digest string
	if wr.WarcHeader().Has(gowarc.WarcPayloadDigest) {
		digest = wr.WarcHeader().Get(gowarc.WarcPayloadDigest)
	} else if wr.WarcHeader().Has(gowarc.WarcBlockDigest) {
		digest = wr.WarcHeader().Get(gowarc.WarcBlockDigest)
	} else {
		if err := wr.Block().Cache(); err != nil {
			panic("Could not cache record: " + err.Error())
		}
		if err := wr.ValidateDigest(validation); err != nil {
			panic("Validate error: " + err.Error())
		}
		return getDigest(wr, validation, result)
	}
	return digest
}

func payloadLength(wr gowarc.WarcRecord) int64 {
	var length int64
	switch v := wr.Block().(type) {
	case gowarc.ProtocolHeaderBlock:
		length, _ = wr.ContentLength()
		length -= int64(len(v.ProtocolHeaderBytes()))
	case gowarc.WarcFieldsBlock:
		length = v.Size()
	}
	return length
}

func revisitRefSize(r *gowarc.RevisitRef) int {
	s := 0
	if r.TargetRecordId != "" {
		s += len(gowarc.WarcRefersTo) + len(r.TargetRecordId)
	}
	if r.Profile != "" {
		s += len(gowarc.WarcProfile) + len(r.Profile)
	}
	if r.TargetUri != "" {
		s += len(gowarc.WarcRefersToTargetURI) + len(r.TargetUri)
	}
	if r.TargetDate != "" {
		s += len(gowarc.WarcRefersToDate) + len(r.TargetDate)
	}
	return s
}
