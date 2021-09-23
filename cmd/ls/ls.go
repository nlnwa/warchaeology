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
	"errors"
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal"
	"github.com/spf13/cobra"
	"io"
	"os"
	"sort"
	"strconv"
)

type conf struct {
	offset      int64
	recordCount int
	strict      bool
	fileName    string
	id          []string
	format      string
	writer      RecordWriter
}

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:   "ls",
		Short: "List records from warc files",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing file name")
			}
			c.fileName = args[0]
			if c.offset >= 0 && c.recordCount == 0 {
				c.recordCount = 1
			}
			if c.offset < 0 {
				c.offset = 0
			}
			sort.Strings(c.id)
			return runE(c)
		},
	}

	cmd.Flags().Int64VarP(&c.offset, "offset", "o", -1, "record offset")
	cmd.Flags().IntVarP(&c.recordCount, "record-count", "c", 0, "The maximum number of records to show")
	cmd.Flags().BoolVarP(&c.strict, "strict", "s", false, "strict parsing")
	cmd.Flags().StringArrayVar(&c.id, "id", []string{}, "specify record ids to ls")
	cmd.Flags().StringVar(&c.format, "format", "", "specify output format. One of: 'cdx', 'cdxj'")

	return cmd
}

func runE(c *conf) error {
	readFile(c, c.fileName)
	return nil
}

func readFile(c *conf, fileName string) {
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
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v, rec num: %v, Offset %v\n", err.Error(), strconv.Itoa(count), currentOffset)
			break
		}
		if len(c.id) > 0 {
			if !internal.Contains(c.id, wr.WarcHeader().Get(gowarc.WarcRecordID)) {
				continue
			}
		}
		count++

		if err := c.writer.Write(wr, fileName, currentOffset); err != nil {
			panic(err)
		}

		if c.recordCount > 0 && count >= c.recordCount {
			break
		}
	}
	fmt.Fprintln(os.Stderr, "Count: ", count)
}
