/*
 * Copyright 2021 National Library of Norway.
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
 */

package arc

import (
	"context"
	"errors"
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/arcreader"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/warcwriterconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type conf struct {
	files       []string
	concurrency int
	writerConf  *warcwriterconfig.WarcWriterConfig
}

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:   "arc <files/dirs>",
		Short: "Convert arc file into warc file",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if wc, err := warcwriterconfig.NewFromViper(); err != nil {
				return err
			} else {
				c.writerConf = wc
			}
			c.concurrency = viper.GetInt(flag.Concurrency)

			if len(args) == 0 {
				return errors.New("missing file or directory name")
			}
			c.files = args
			return runE(cmd.Name(), c)
		},
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
	cmd.Flags().StringP(flag.FilePrefix, "p", "", flag.FilePrefixHelp)
	cmd.Flags().StringP(flag.WarcDir, "w", ".", flag.WarcDirHelp)
	cmd.Flags().String(flag.SubdirPattern, "", flag.SubdirPatternHelp)
	cmd.Flags().StringP(flag.NameGenerator, "n", "default", flag.NameGeneratorHelp)
	cmd.Flags().Bool(flag.Flush, false, flag.FlushHelp)
	cmd.Flags().String(flag.WarcVersion, "1.1", flag.WarcVersionHelp)
	cmd.Flags().StringP(flag.DefaultDate, "t", time.Now().Format(warcwriterconfig.DefaultDateFormat), flag.DefaultDateHelp)

	return cmd
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
	fileWalker := filewalker.NewFromViper(cmd, c.files, c.readFile)
	stats := filewalker.NewStats()
	return fileWalker.Walk(ctx, stats)
}

func (c *conf) readFile(fileName string) filewalker.Result {
	result := filewalker.NewResult(fileName)

	a, err := arcreader.NewArcFileReader(fileName, 0,
		gowarc.WithVersion(c.writerConf.WarcVersion),
		gowarc.WithAddMissingDigest(true),
		gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)),
	)
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
		var currentOffset int64
		currentOffset, writer, err = handleRecord(c, a, fileName, result, writer)
		if err == io.EOF {
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
func handleRecord(c *conf, wf *arcreader.ArcFileReader, fileName string, result filewalker.Result, writer *gowarc.WarcFileWriter) (offset int64, writerOut *gowarc.WarcFileWriter, err error) {
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
