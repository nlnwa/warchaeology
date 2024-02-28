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

	if wc, err := warcwriterconfig.NewFromViper(cmd.Name()); err != nil {
		return err
	} else {
		wc.WarcInfoFunc = func(recordBuilder gowarc.WarcRecordBuilder) error {
			payload := &gowarc.WarcFields{}
			payload.Set("software", cmdversion.SoftwareVersion()+" https://github.com/nlnwa/warchaeology")
			payload.Set("format", fmt.Sprintf("WARC File Format %d.%d", wc.WarcVersion.Minor(), wc.WarcVersion.Minor()))
			payload.Set("description", "Converted from Nedlib")
			h, e := os.Hostname()
			if e != nil {
				return e
			}
			payload.Set("host", h)

			_, err := recordBuilder.WriteString(payload.String())
			return err
		}

		config.writerConf = wc
	}
	config.concurrency = viper.GetInt(flag.Concurrency)

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

	r, err := nedlibreader.NewNedlibReader(fs, fileName, c.writerConf.DefaultTime,
		gowarc.WithVersion(c.writerConf.WarcVersion),
		gowarc.WithAddMissingDigest(true),
		gowarc.WithFixDigest(true),
		gowarc.WithFixContentLength(true),
		gowarc.WithAddMissingContentLength(true),
	)
	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = r.Close() }()

	_, err = handleRecord(c, r, fileName, result)
	if err != nil {
		result.AddError(fmt.Errorf("error: %v, rec num: %d", err.Error(), result.Records()))
	}
	return result
}

// handleRecord processes one record
func handleRecord(c *conf, wf *nedlibreader.NedlibReader, fileName string, result filewalker.Result) (offset int64, err error) {
	wr, currentOffset, validation, e := wf.Next()
	offset = currentOffset
	if e != nil {
		err = e
		return
	}
	result.IncrRecords()
	result.IncrProcessed()
	if !validation.Valid() {
		result.AddError(fmt.Errorf("info: found problem in rec num: %d, offset %d: %s", result.Records(), currentOffset, validation))
	}

	defer func() { _ = wr.Close() }()

	syntheticFileName, err := internal.To14(wr.WarcHeader().Get(gowarc.WarcDate))
	if err != nil {
		panic(err)
	}

	writer := c.writerConf.GetWarcWriter(syntheticFileName, wr.WarcHeader().Get(gowarc.WarcDate))

	if rr := writer.Write(wr); rr != nil && rr[0].Err != nil {
		fmt.Printf("Offset: %d\n", currentOffset)
		_, _ = wr.WarcHeader().Write(os.Stdout)
		panic(rr[0].Err)
	}
	return
}
