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
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/filter"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type conf struct {
	files       []string
	offset      int64
	recordCount int
	strict      bool
	format      string
	fields      string
	filter      *filter.Filter
	delimiter   string
	writer      *RecordWriter
	concurrency int
}

func NewCommand() *cobra.Command {
	c := &conf{}
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing file or directory")
			}
			c.files = args
			if c.offset >= 0 && c.recordCount == 0 {
				c.recordCount = 1
				// TODO: check that input is exactly one file when using offset
			}
			if c.offset < 0 {
				c.offset = 0
			}

			if !cmd.Flag(flag.LogConsole).Changed {
				viper.Set(flag.LogConsole, []string{"summary"})
			}

			c.filter = filter.NewFromViper()

			return runE(cmd.Name(), c)
		},
		ValidArgsFunction: flag.SuffixCompletionFn,
	}

	cmd.Flags().BoolP(flag.Recursive, "r", false, flag.RecursiveHelp)
	cmd.Flags().BoolP(flag.FollowSymlinks, "s", false, flag.FollowSymlinksHelp)
	cmd.Flags().StringSlice(flag.Suffixes, []string{".warc", ".warc.gz"}, flag.SuffixesHelp)
	cmd.Flags().IntVarP(&c.concurrency, flag.Concurrency, "c", 1, flag.ConcurrencyHelp)
	cmd.Flags().Int64VarP(&c.offset, "offset", "o", -1, "record offset")
	cmd.Flags().IntVarP(&c.recordCount, "record-count", "n", 0, "The maximum number of records to show")
	cmd.Flags().BoolVar(&c.strict, "strict", false, "strict parsing")
	cmd.Flags().StringVarP(&c.delimiter, "delimiter", "d", " ", "use string instead of SPACE for field delimiter")
	cmd.Flags().StringVarP(&c.fields, "fields", "f", "", "which fields to include. See 'warc help ls' for a description")
	cmd.Flags().StringArray(flag.RecordId, []string{}, flag.RecordIdHelp)
	cmd.Flags().StringSliceP(flag.RecordType, "t", []string{}, flag.RecordTypeHelp)
	cmd.Flags().StringP(flag.ResponseCode, "S", "", flag.ResponseCodeHelp)
	cmd.Flags().StringSliceP(flag.MimeType, "m", []string{}, flag.MimeTypeHelp)

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

func runE(cmd string, c *conf) error {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	fileWalker := filewalker.NewFromViper(cmd, c.files, c.readFile)
	res := filewalker.NewStats()
	return fileWalker.Walk(ctx, res)
}

func (c *conf) readFile(fileName string) filewalker.Result {
	result := filewalker.NewResult(fileName)

	var opts []gowarc.WarcRecordOption
	opts = append(opts, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	if c.strict {
		opts = append(opts, gowarc.WithStrictValidation())
	} else {
		opts = append(opts, gowarc.WithSyntaxErrorPolicy(gowarc.ErrIgnore), gowarc.WithSpecViolationPolicy(gowarc.ErrIgnore), gowarc.WithUnknownRecordTypePolicy(gowarc.ErrIgnore))
	}
	wf, err := gowarc.NewWarcFileReader(fileName, c.offset, opts...)
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
				line = ""
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
