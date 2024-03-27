package nedlib

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal"
	"github.com/nlnwa/warchaeology/internal/cmdversion"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/warcwriterconfig"
	"github.com/nlnwa/warchaeology/nedlibreader"
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
		Use:   "nedlib <files/dirs>",
		Short: "Convert directory with files harvested with Nedlib into warc files",
		Long:  ``,
		RunE:  parseArgumentsAndCallNedlib,
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
	cmd.Flags().StringSlice(flag.Suffixes, []string{".meta"}, flag.SuffixesHelp)
	cmd.Flags().IntP(flag.Concurrency, "c", int(float32(runtime.NumCPU())*float32(1.5)), flag.ConcurrencyHelp)
	cmd.Flags().IntP(flag.ConcurrentWriters, "C", 1, flag.ConcurrentWritersHelp)
	cmd.Flags().Int64P(flag.FileSize, "S", 1024*1024*1024, flag.FileSizeHelp)
	cmd.Flags().BoolP(flag.Compress, "z", false, flag.CompressHelp)
	cmd.Flags().Bool(flag.CompressionLevel, false, flag.CompressionLevelHelp)
	cmd.Flags().StringP(flag.FilePrefix, "p", "from_nedlib_", flag.FilePrefixHelp)
	cmd.Flags().StringP(flag.WarcDir, "w", ".", flag.WarcDirHelp)
	cmd.Flags().String(flag.SubdirPattern, "", flag.SubdirPatternHelp)
	cmd.Flags().Bool(flag.Flush, false, flag.FlushHelp)
	cmd.Flags().String(flag.WarcVersion, "1.1", flag.WarcVersionHelp)
	cmd.Flags().StringP(flag.DefaultDate, "t", time.Now().Format(warcwriterconfig.DefaultDateFormat), flag.DefaultDateHelp)
	cmd.Flags().String(flag.SrcFilesystem, "", flag.SrcFilesystemHelp)
	cmd.Flags().String(flag.OpenInputFileHook, "", flag.OpenInputFileHookHelp)
	cmd.Flags().String(flag.CloseInputFileHook, "", flag.CloseInputFileHookHelp)
	cmd.Flags().String(flag.OpenOutputFileHook, "", flag.OpenOutputFileHookHelp)
	cmd.Flags().String(flag.CloseOutputFileHook, "", flag.CloseOutputFileHookHelp)

	return cmd
}

func parseArgumentsAndCallNedlib(cmd *cobra.Command, args []string) error {
	config := &conf{}
	// The Nedlib data structure does not support direct filename transformations.
	// Instead, we employ a custom generator that treats the input filename as a date.
	// When we request a new warcwriter, we submit a synthetic fromFilename based on the date of the first record.
	viper.Set(flag.NameGenerator, "nedlib")

	if warcWriterConfig, err := warcwriterconfig.NewFromViper(cmd.Name()); err != nil {
		return err
	} else {
		warcWriterConfig.WarcInfoFunc = func(recordBuilder gowarc.WarcRecordBuilder) error {
			payload := &gowarc.WarcFields{}
			payload.Set("software", cmdversion.SoftwareVersion()+" https://github.com/nlnwa/warchaeology")
			payload.Set("format", fmt.Sprintf("WARC File Format %d.%d", warcWriterConfig.WarcVersion.Minor(), warcWriterConfig.WarcVersion.Minor()))
			payload.Set("description", "Converted from Nedlib")
			hostname, errInner := os.Hostname()
			if errInner != nil {
				return errInner
			}
			payload.Set("host", hostname)

			_, err := recordBuilder.WriteString(payload.String())
			return err
		}

		config.writerConf = warcWriterConfig
	}
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

	nedlibReader, err := nedlibreader.NewNedlibReader(fileSystem, fileName, config.writerConf.DefaultTime,
		gowarc.WithVersion(config.writerConf.WarcVersion),
		gowarc.WithAddMissingDigest(true),
		gowarc.WithFixDigest(true),
		gowarc.WithFixContentLength(true),
		gowarc.WithAddMissingContentLength(true),
	)
	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = nedlibReader.Close() }()

	_, err = handleRecord(config, nedlibReader, fileName, result)
	if err != nil {
		result.AddError(fmt.Errorf("error: %v, rec num: %d", err.Error(), result.Records()))
	}
	return result
}

// handleRecord processes one record
func handleRecord(config *conf, nedlibReader *nedlibreader.NedlibReader, fileName string, result filewalker.Result) (int64, error) {
	warcRecord, offset, validation, err := nedlibReader.Next()
	if err != nil {
		return offset, err
	}
	result.IncrRecords()
	result.IncrProcessed()
	if !validation.Valid() {
		result.AddError(fmt.Errorf("info: found problem in rec num: %d, offset %d: %s", result.Records(), offset, validation))
	}

	defer func() { _ = warcRecord.Close() }()

	syntheticFileName, err := internal.To14(warcRecord.WarcHeader().Get(gowarc.WarcDate))
	if err != nil {
		panic(err)
	}

	writer := config.writerConf.GetWarcWriter(syntheticFileName, warcRecord.WarcHeader().Get(gowarc.WarcDate))

	if writeResponse := writer.Write(warcRecord); writeResponse != nil && writeResponse[0].Err != nil {
		fmt.Printf("Offset: %d\n", offset)
		_, _ = warcRecord.WarcHeader().Write(os.Stdout)
		panic(writeResponse[0].Err)
	}
	return offset, err
}
