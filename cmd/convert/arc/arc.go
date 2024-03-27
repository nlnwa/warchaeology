package arc

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
	"github.com/nlnwa/warchaeology/arcreader"
	"github.com/nlnwa/warchaeology/internal/cmdversion"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/warcwriterconfig"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type conf struct {
	files       []string
	concurrency int
	writerConf  *warcwriterconfig.WarcWriterConfig
}

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "arc <files/dirs>",
		Short:             "Convert arc file into warc file",
		Long:              ``,
		RunE:              parseArgumentsAndCallArc,
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
	cmd.Flags().StringSlice(flag.Suffixes, []string{".arc", ".arc.gz"}, flag.SuffixesHelp)
	cmd.Flags().IntP(flag.Concurrency, "c", int(float32(runtime.NumCPU())*float32(1.5)), flag.ConcurrencyHelp)
	cmd.Flags().IntP(flag.ConcurrentWriters, "C", 1, flag.ConcurrentWritersHelp)
	cmd.Flags().Int64P(flag.FileSize, "S", 1024*1024*1024, flag.FileSizeHelp)
	cmd.Flags().BoolP(flag.Compress, "z", false, flag.CompressHelp)
	cmd.Flags().Bool(flag.CompressionLevel, false, flag.CompressionLevelHelp)
	cmd.Flags().StringP(flag.FilePrefix, "p", "from_arc_", flag.FilePrefixHelp)
	cmd.Flags().StringP(flag.WarcDir, "w", ".", flag.WarcDirHelp)
	cmd.Flags().String(flag.SubdirPattern, "", flag.SubdirPatternHelp)
	cmd.Flags().StringP(flag.NameGenerator, "n", "identity", flag.NameGeneratorHelp)
	cmd.Flags().Bool(flag.Flush, false, flag.FlushHelp)
	cmd.Flags().String(flag.WarcVersion, "1.1", flag.WarcVersionHelp)
	cmd.Flags().StringP(flag.DefaultDate, "t", time.Now().Format(warcwriterconfig.DefaultDateFormat), flag.DefaultDateHelp)
	cmd.Flags().String(flag.SrcFilesystem, "", flag.SrcFilesystemHelp)
	cmd.Flags().String(flag.OpenInputFileHook, "", flag.OpenInputFileHookHelp)
	cmd.Flags().String(flag.CloseInputFileHook, "", flag.CloseInputFileHookHelp)
	cmd.Flags().String(flag.OpenOutputFileHook, "", flag.OpenOutputFileHookHelp)
	cmd.Flags().String(flag.CloseOutputFileHook, "", flag.CloseOutputFileHookHelp)

	if err := cmd.RegisterFlagCompletionFunc(flag.NameGenerator,
		cobra.FixedCompletions([]string{"default", "identity"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	return cmd
}

func parseArgumentsAndCallArc(cmd *cobra.Command, args []string) error {
	config := &conf{}
	warcRecordWriter, err := warcwriterconfig.NewFromViper(cmd.Name())
	if err != nil {
		return err
	}

	warcRecordWriter.OneToOneWriter = true

	if warcRecordWriter.OneToOneWriter {
		warcRecordWriter.WarcInfoFunc = func(recordBuilder gowarc.WarcRecordBuilder) error {
			payload := &gowarc.WarcFields{}
			payload.Set("software", cmdversion.SoftwareVersion()+" https://github.com/nlnwa/warchaeology")
			payload.Set("format", fmt.Sprintf("WARC File Format %d.%d", warcRecordWriter.WarcVersion.Minor(), warcRecordWriter.WarcVersion.Minor()))
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

	config.writerConf = warcRecordWriter
	config.concurrency = viper.GetInt(flag.Concurrency)

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

	arcFileReader, err := arcreader.NewArcFileReader(fileSystem, fileName, 0,
		gowarc.WithVersion(config.writerConf.WarcVersion),
		gowarc.WithAddMissingDigest(true),
		gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)),
	)
	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = arcFileReader.Close() }()

	var warcFileWriter *gowarc.WarcFileWriter

	for {
		var currentOffset int64
		currentOffset, warcFileWriter, err = handleRecord(config, arcFileReader, fileName, result, warcFileWriter)
		if err == io.EOF {
			break
		}
		if err != nil {
			result.AddError(fmt.Errorf("error: %v, rec num: %d, offset %d", err.Error(), result.Records(), currentOffset))
		}
	}

	if warcFileWriter != nil {
		_ = warcFileWriter.Close()
	}

	return result
}

// handleRecord processes one record
// The input parameter writer and output parameter writerOut is only used for identity transformation. In this case there is one writer
// per file which should be closed by readFile when the file is processed. But since we need the warc date of the first record to open
// the writer, it must be opened in this function. These parameters are used for giving readFile access to the writer.
func handleRecord(config *conf, arcFileReader *arcreader.ArcFileReader, fileName string, result filewalker.Result, warcFileWriter *gowarc.WarcFileWriter) (int64, *gowarc.WarcFileWriter, error) {
	writerOut := warcFileWriter

	warcRecord, offset, validation, err := arcFileReader.Next()
	defer func() {
		if warcRecord != nil {
			_ = warcRecord.Close()
		}
	}()

	result.IncrRecords()
	result.IncrProcessed()
	if err != nil {
		return offset, writerOut, err
	}
	if !validation.Valid() {
		fmt.Println(fmt.Errorf("info: found problem in rec num: %d, offset %d: %s", result.Records(), offset, validation))
		result.AddError(fmt.Errorf("info: found problem in rec num: %d, offset %d: %s", result.Records(), offset, validation))
	}

	if warcFileWriter == nil {
		warcFileWriter = config.writerConf.GetWarcWriter(fileName, warcRecord.WarcHeader().Get(gowarc.WarcDate))
		if config.writerConf.OneToOneWriter {
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
