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
	"github.com/nlnwa/warchaeology/internal"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
)

type conf struct {
	files       []string
	offset      int64
	recordCount int
	strict      bool
	id          []string
	format      string
	writer      RecordWriter
	concurrency int
}

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:   "ls <files/dirs>",
		Short: "List records from warc files",
		Long:  ``,
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
			sort.Strings(c.id)

			if !cmd.Flag(flag.LogConsole).Changed {
				viper.Set(flag.LogConsole, []string{"summary"})
			}
			return runE(c)
		},
	}

	cmd.Flags().BoolP(flag.Recursive, "r", false, flag.RecursiveHelp)
	cmd.Flags().BoolP(flag.FollowSymlinks, "s", false, flag.FollowSymlinksHelp)
	cmd.Flags().StringSlice(flag.Suffixes, []string{".warc", ".warc.gz"}, flag.SuffixesHelp)
	cmd.Flags().IntVarP(&c.concurrency, flag.Concurrency, "c", 1, flag.ConcurrencyHelp)
	cmd.Flags().Int64VarP(&c.offset, "offset", "o", -1, "record offset")
	cmd.Flags().IntVarP(&c.recordCount, "record-count", "n", 0, "The maximum number of records to show")
	cmd.Flags().BoolVar(&c.strict, "strict", false, "strict parsing")
	cmd.Flags().StringArrayVar(&c.id, "id", []string{}, "specify record ids to ls")
	cmd.Flags().StringVar(&c.format, "format", "", "specify output format. One of: 'cdx', 'cdxj'")

	return cmd
}

func runE(c *conf) error {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	fileWalker := filewalker.NewFromViper(c.files, c.readFile)
	res := filewalker.NewStats()
	return fileWalker.Walk(ctx, res)
}

func (c *conf) readFile(fileName string) filewalker.Result {
	result := filewalker.NewResult(fileName)

	var opts []gowarc.WarcRecordOption
	if c.strict {
		opts = append(opts, gowarc.WithStrictValidation())
	} else {
		opts = append(opts, gowarc.WithNoValidation())
	}
	wf, err := gowarc.NewWarcFileReader(fileName, c.offset, opts...)
	defer func() { _ = wf.Close() }()
	if err != nil {
		panic(err)
	}

	if c.format != "" {
		switch c.format {
		case "cdx":
			c.writer = &CdxLegacy{}
		case "cdxj":
			c.writer = &CdxJ{}
		default:
			panic(fmt.Errorf("unknwon format %v, valid formats are: 'cdx', 'cdxj'", c.format))
		}
	} else {
		c.writer = &DefaultWriter{}
	}

	count := 0

	for {
		wr, currentOffset, _, err := wf.Next()
		if err == io.EOF {
			break
		}
		result.IncrRecords()
		if err != nil {
			result.AddError(fmt.Errorf("error: %v, rec num: %v, offset %v", err.Error(), strconv.Itoa(count), currentOffset))
			break
		}
		if len(c.id) > 0 {
			if !internal.Contains(c.id, wr.WarcHeader().Get(gowarc.WarcRecordID)) {
				continue
			}
		}
		count++

		result.IncrProcessed()
		if err := c.writer.Write(wr, fileName, currentOffset); err != nil {
			panic(err)
		}

		if c.recordCount > 0 && count >= c.recordCount {
			break
		}
	}
	return result
}
