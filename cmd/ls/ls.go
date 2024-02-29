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

package ls

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/filter"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type conf struct {
	files       []string
	offset      int64
	recordCount int
	strict      bool
	fields      string
	filter      *filter.Filter
	delimiter   string
	writer      *RecordWriter
	concurrency int
}

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ls <files/dirs>",
		Short: "List warc file contents",
		Long: `List information about records in one or more warc files.

Output options:

    --delimiter accepts a string to be used as the output field delimiter.
    --fields specifies which fields to include in output. Field specification letters are mostly the same as the fields in
           the CDX file specification (https://iipc.github.io/warc-specifications/specifications/cdx-format/cdx-2015/).
           The following fields are supported:
             a - original URL
             b - date in 14 digit format
             B - date in RFC3339 format
             e - IP
             g - file name
             h - original host
             i - record id
             k - checksum
             m - document mime type
             s - http response code
             S - record size in WARC file
             T - record type
             V - Offset in WARC file
           A number after the field letter restricts the field length. By adding a + or - sign before the number the field is
           padded to have the exact length. + is right aligned and - is left aligned.`,
		RunE:              parseArgumentsAndCallLs,
		ValidArgsFunction: flag.SuffixCompletionFn,
	}

	cmd.Flags().BoolP(flag.Recursive, "r", false, flag.RecursiveHelp)
	cmd.Flags().BoolP(flag.FollowSymlinks, "s", false, flag.FollowSymlinksHelp)
	cmd.Flags().StringSlice(flag.Suffixes, []string{".warc", ".warc.gz"}, flag.SuffixesHelp)
	cmd.Flags().IntP(flag.Concurrency, "c", 1, flag.ConcurrencyHelp)
	cmd.Flags().Int64P(flag.Offset, "o", -1, flag.OffsetHelp)
	cmd.Flags().IntP(flag.RecordCount, "n", 0, flag.RecordCountHelp)
	cmd.Flags().Bool(flag.Strict, false, flag.StrictHelp)
	cmd.Flags().StringP(flag.Delimiter, "d", " ", flag.DelimiterHelp)
	cmd.Flags().StringP(flag.Fields, "f", "", flag.FieldsHelp)
	cmd.Flags().StringArray(flag.RecordId, []string{}, flag.RecordIdHelp)
	cmd.Flags().StringSliceP(flag.RecordType, "t", []string{}, flag.RecordTypeHelp)
	cmd.Flags().StringP(flag.ResponseCode, "S", "", flag.ResponseCodeHelp)
	cmd.Flags().StringSliceP(flag.MimeType, "m", []string{}, flag.MimeTypeHelp)
	cmd.Flags().String(flag.SrcFilesystem, "", flag.SrcFilesystemHelp)
	cmd.Flags().String(flag.SrcFileList, "", flag.SrcFileListHelp)

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

	return cmd
}

func parseArgumentsAndCallLs(cmd *cobra.Command, args []string) error {
	config := &conf{}
	if len(args) == 0 && viper.GetString(flag.SrcFileList) == "" {
		return errors.New("missing file or directory name")
	}
	config.files = args
	config.delimiter = viper.GetString(flag.Delimiter)
	config.concurrency = viper.GetInt(flag.Concurrency)
	config.offset = viper.GetInt64(flag.Offset)
	config.recordCount = viper.GetInt(flag.RecordCount)
	config.strict = viper.GetBool(flag.Strict)
	config.fields = viper.GetString(flag.Fields)

	if config.offset >= 0 && config.recordCount == 0 {
		config.recordCount = 1
		// TODO: check that input is exactly one file when using offset
	}
	if config.offset < 0 {
		config.offset = 0
	}

	if !cmd.Flag(flag.LogConsole).Changed {
		viper.Set(flag.LogConsole, []string{"summary"})
	}

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

	fileWalker, err := filewalker.NewFromViper(cmd, c.files, c.readFile)
	if err != nil {
		return err
	}
	res := filewalker.NewStats()
	return fileWalker.Walk(ctx, res)
}

func (c *conf) readFile(fs afero.Fs, fileName string) filewalker.Result {
	result := filewalker.NewResult(fileName)

	var opts []gowarc.WarcRecordOption
	opts = append(opts, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if c.strict {
		opts = append(opts, gowarc.WithStrictValidation())
	} else {
		opts = append(opts, gowarc.WithSyntaxErrorPolicy(gowarc.ErrIgnore), gowarc.WithSpecViolationPolicy(gowarc.ErrIgnore), gowarc.WithUnknownRecordTypePolicy(gowarc.ErrIgnore))
	}

	f, err := fs.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()
	wf, err := gowarc.NewWarcFileReaderFromStream(f, c.offset, opts...)
	defer func() { _ = wf.Close() }()
	if err != nil {
		panic(err)
	}

	if c.fields == "" {
		c.fields = "V+11iT-8a100"
	}
	c.writer = NewRecordWriter(c.fields, c.delimiter)

	count := 0
	var lastOffset int64

	var line string
	for {
		wr, currentOffset, _, err := wf.Next()
		size := currentOffset - lastOffset
		lastOffset = currentOffset

		if err == io.EOF || (c.recordCount > 0 && count >= c.recordCount) {
			if line != "" {
				c.writer.Write(line, size)
			}
			break
		}

		if err != nil {
			result.AddError(fmt.Errorf("error: %v, rec num: %v, offset %v", err.Error(), strconv.Itoa(count), currentOffset))
			break
		}

		if line != "" {
			c.writer.Write(line, size)
			line = ""
		}

		result.IncrRecords()
		if !c.filter.Accept(wr) {
			continue
		}

		count++

		result.IncrProcessed()
		line = c.writer.FormatRecord(wr, fileName, currentOffset)
	}
	return result
}
