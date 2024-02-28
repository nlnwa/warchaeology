/*
 * Copyright 2022 National Library of Norway.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
	if wc, err := warcwriterconfig.NewFromViper(cmd.Name()); err != nil {
		return err
	} else {
		config.writerConf = wc
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

func runE(cmd string, c *conf) error {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	defer c.writerConf.Close()
	fileWalker, err := filewalker.NewFromViper(cmd, c.files, c.readFile)
	if err != nil {
		return err
	}
	stats := filewalker.NewStats()
	return fileWalker.Walk(ctx, stats)
}

func (c *conf) readFile(fs afero.Fs, fileName string) filewalker.Result {
	result := filewalker.NewResult(fileName)

	opts := []gowarc.WarcRecordOption{
		gowarc.WithVersion(c.writerConf.WarcVersion),
		gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)),
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
			gowarc.WithAddMissingDigest(false),
			gowarc.WithFixSyntaxErrors(false),
			gowarc.WithFixDigest(false),
			gowarc.WithAddMissingContentLength(false),
			gowarc.WithAddMissingRecordId(false),
			gowarc.WithFixContentLength(false),
		)
	}

	f, err := fs.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()
	a, err := gowarc.NewWarcFileReaderFromStream(f, 0, opts...)
	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = a.Close() }()

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
		currentOffset, writer, err = handleRecord(c, a, fileName, result, writer)
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
func handleRecord(c *conf, wf *gowarc.WarcFileReader, fileName string, result filewalker.Result, writer *gowarc.WarcFileWriter) (offset int64, writerOut *gowarc.WarcFileWriter, err error) {
	writerOut = writer

	wr, currentOffset, validation, e := wf.Next()
	offset = currentOffset
	result.IncrRecords()
	result.IncrProcessed()
	defer func() {
		if wr != nil {
			_ = wr.Close()
		}
	}()
	if e != nil {
		err = e
		return
	}
	if !validation.Valid() {
		result.AddError(fmt.Errorf("info: found problem in rec num: %d, offset %d: %s", result.Records(), currentOffset, validation))
		fmt.Printf("%T -- %s\n", wr.Block(), validation)
		rb := gowarc.NewRecordBuilder(wr.Type(), gowarc.WithFixContentLength(false), gowarc.WithFixDigest(false))
		wf := wr.WarcHeader()
		for _, f := range *wf {
			if f.Name != gowarc.WarcType {
				rb.AddWarcHeader(f.Name, f.Value)
			}
		}
		r, err := wr.Block().RawBytes()
		if err != nil {
			panic(err)
		}
		_, err = rb.ReadFrom(r)
		if err != nil {
			panic(err)
		}
		wr, _, err = rb.Build()
		if err != nil {
			panic(err)
		}
	}

	if writer == nil {
		writer = c.writerConf.GetWarcWriter(fileName, wr.WarcHeader().Get(gowarc.WarcDate))
		if c.writerConf.WarcFileNameGenerator == "identity" {
			writerOut = writer
		}
	}
	if rr := writer.Write(wr); rr != nil && rr[0].Err != nil {
		fmt.Printf("Offset: %d\n", currentOffset)
		_, _ = wr.WarcHeader().Write(os.Stdout)
		panic(rr[0].Err)
	}
	return
}
