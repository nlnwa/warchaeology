package warc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/utils"
	"github.com/nlnwa/warchaeology/internal/warcwriterconfig"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type conf struct {
	files           []string
	concurrency     int
	writerConf      *warcwriterconfig.WarcWriterConfig
	minWARCDiskFree int64
	repair          bool
}

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "warc <files/dirs>",
		Short: "Convert WARC file into WARC file",
		Long: `The WARC to WARC converter can be used to reorganize, convert or repair WARC-records.
This is an experimental feature.`,
		RunE:              parseArgumentsAndCallWarc,
		ValidArgsFunction: flag.SuffixCompletionFn,
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	cmd.Flags().BoolP(flag.Recursive, "r", false, flag.RecursiveHelp)
	cmd.Flags().BoolP(flag.FollowSymlinks, "s", false, flag.FollowSymlinksHelp)
	cmd.Flags().BoolP(flag.KeepIndex, "k", false, flag.KeepIndexHelp)
	cmd.Flags().BoolP(flag.NewIndex, "K", false, flag.NewIndexHelp)
	cmd.Flags().StringP(flag.IndexDir, "i", cacheDir+"/warc", flag.IndexDirHelp)
	cmd.Flags().StringSlice(flag.Suffixes, []string{".warc", ".warc.gz"}, flag.SuffixesHelp)
	cmd.Flags().IntP(flag.Concurrency, "c", int(float32(runtime.NumCPU())*float32(1.5)), flag.ConcurrencyHelp)
	cmd.Flags().IntP(flag.ConcurrentWriters, "C", 1, flag.ConcurrentWritersHelp)
	cmd.Flags().Int64P(flag.FileSize, "S", 1024*1024*1024, flag.FileSizeHelp)
	cmd.Flags().BoolP(flag.Compress, "z", false, flag.CompressHelp)
	cmd.Flags().Bool(flag.CompressionLevel, false, flag.CompressionLevelHelp)
	cmd.Flags().StringP(flag.FilePrefix, "p", "", flag.FilePrefixHelp)
	cmd.Flags().StringP(flag.WarcDir, "w", ".", flag.WarcDirHelp)
	cmd.Flags().String(flag.SubdirPattern, "", flag.SubdirPatternHelp)
	cmd.Flags().StringP(flag.NameGenerator, "n", "default", flag.NameGeneratorHelp)
	cmd.Flags().Bool(flag.Flush, false, flag.FlushHelp)
	cmd.Flags().String(flag.WarcVersion, "1.1", flag.WarcVersionHelp)
	cmd.Flags().StringP(flag.DefaultDate, "t", time.Now().Format(warcwriterconfig.DefaultDateFormat), flag.DefaultDateHelp)
	cmd.Flags().String(flag.MinFreeDisk, "256MB", flag.MinFreeDiskHelp)
	cmd.Flags().BoolP(flag.Repair, "R", false, flag.RepairHelp)
	cmd.Flags().String(flag.SrcFilesystem, "", flag.SrcFilesystemHelp)
	cmd.Flags().String(flag.OpenInputFileHook, "", flag.OpenInputFileHookHelp)
	cmd.Flags().String(flag.CloseInputFileHook, "", flag.CloseInputFileHookHelp)
	cmd.Flags().String(flag.OpenOutputFileHook, "", flag.OpenOutputFileHookHelp)
	cmd.Flags().String(flag.CloseOutputFileHook, "", flag.CloseOutputFileHookHelp)

	return cmd
}

func parseArgumentsAndCallWarc(cmd *cobra.Command, args []string) error {
	config := &conf{}
	if warcWriterConfig, err := warcwriterconfig.NewFromViper(cmd.Name()); err != nil {
		return err
	} else {
		config.writerConf = warcWriterConfig
	}
	config.concurrency = viper.GetInt(flag.Concurrency)
	config.minWARCDiskFree = utils.ParseSizeInBytes(viper.GetString(flag.MinFreeDisk))
	config.repair = viper.GetBool(flag.Repair)

	if len(args) == 0 && viper.GetString(flag.SrcFileList) == "" {
		return errors.New("missing file or directory name")
	}
	config.files = args
	return runE(cmd.Name(), config)

}

func runE(cmd string, config *conf) error {
	ctx, cancel := context.WithCancel(context.Background())

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChannel
		cancel()
	}()

	defer config.writerConf.Close()
	fileWalker, err := filewalker.NewFromViper(cmd, config.files, config.readFile)
	if err != nil {
		return err
	}
	stats := filewalker.NewStats()
	return fileWalker.Walk(ctx, stats)
}

func (config *conf) readFile(fileSystem afero.Fs, fileName string) filewalker.Result {
	result := filewalker.NewResult(fileName)

	warcRecordOptions := []gowarc.WarcRecordOption{
		gowarc.WithVersion(config.writerConf.WarcVersion),
		gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)),
	}
	if config.repair {
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

	file, err := fileSystem.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer func() { _ = file.Close() }()
	warcFileReader, err := gowarc.NewWarcFileReaderFromStream(file, 0, warcRecordOptions...)
	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = warcFileReader.Close() }()

	var warcFileWriter *gowarc.WarcFileWriter
	if config.writerConf.WarcFileNameGenerator == "identity" {
		defer func() {
			if warcFileWriter != nil {
				_ = warcFileWriter.Close()
			}
		}()
	}

	for {
		if utils.DiskFree(config.writerConf.OutDir) < config.minWARCDiskFree {
			result.SetFatal(utils.NewOutOfSpaceError("cannot write WARC file, almost no space left in directory '%s'\n", config.writerConf.OutDir))
			break
		}
		var currentOffset int64
		currentOffset, warcFileWriter, err = handleRecord(config, warcFileReader, fileName, result, warcFileWriter)
		if err == io.EOF {
			break
		}
		if result.Fatal() != nil {
			break
		}
		if err != nil {
			result.AddError(fmt.Errorf("error: %v, rec num: %d, offset %d", err.Error(), result.Records(), currentOffset))
		}
	}

	return result
}

// handleRecord processes one record
// The input parameter writer and output parameter writerOut is only used for identity transformation. In this case there is one writer
// per file which should be closed by readFile when the file is processed. But since we need the warc date of the first record to open
// the writer, it must be opened in this function. These parameters are used for giving readFile access to the writer.
func handleRecord(config *conf, warcFileReader *gowarc.WarcFileReader, fileName string, result filewalker.Result, warcFileWriter *gowarc.WarcFileWriter) (int64, *gowarc.WarcFileWriter, error) {
	writerOut := warcFileWriter

	warcRecord, offset, validation, err := warcFileReader.Next()
	result.IncrRecords()
	result.IncrProcessed()
	defer func() {
		if warcRecord != nil {
			_ = warcRecord.Close()
		}
	}()
	if err != nil {
		return offset, writerOut, err
	}
	if !validation.Valid() {
		result.AddError(fmt.Errorf("info: found problem in rec num: %d, offset %d: %s", result.Records(), offset, validation))
		fmt.Printf("%T -- %s\n", warcRecord.Block(), validation)
		warcRecordBuilder := gowarc.NewRecordBuilder(warcRecord.Type(), gowarc.WithFixContentLength(false), gowarc.WithFixDigest(false))
		warcFields := warcRecord.WarcHeader()
		for _, warcField := range *warcFields {
			if warcField.Name != gowarc.WarcType {
				warcRecordBuilder.AddWarcHeader(warcField.Name, warcField.Value)
			}
		}
		ioReader, err := warcRecord.Block().RawBytes()
		if err != nil {
			panic(err)
		}
		_, err = warcRecordBuilder.ReadFrom(ioReader)
		if err != nil {
			panic(err)
		}
		warcRecord, _, err = warcRecordBuilder.Build()
		if err != nil {
			panic(err)
		}
	}

	if warcFileWriter == nil {
		warcFileWriter = config.writerConf.GetWarcWriter(fileName, warcRecord.WarcHeader().Get(gowarc.WarcDate))
		if config.writerConf.WarcFileNameGenerator == "identity" {
			writerOut = warcFileWriter
		}
	}
	if writeResponse := warcFileWriter.Write(warcRecord); writeResponse != nil && writeResponse[0].Err != nil {
		fmt.Printf("Offset: %d\n", offset)
		_, _ = warcRecord.WarcHeader().Write(os.Stdout)
		panic(writeResponse[0].Err)
	}
	return offset, writerOut, err
}
