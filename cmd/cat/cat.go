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

package cat

import (
	"errors"
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filter"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/viper"
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

type conf struct {
	offset             int64
	recordNum          int
	recordCount        int
	fileName           string
	filter             *filter.Filter
	showWarcHeader     bool
	showProtocolHeader bool
	showPayload        bool
}

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:   "cat",
		Short: "Concatenate and print warc files",
		Long:  ``,
		Example: `# Print all content from a WARC file
warc cat file1.warc.gz

# Pipe payload from record #4 into the image viewer feh
warc cat -n4 -P file1.warc.gz | feh -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing file name")
			}
			c.fileName = args[0]
			c.offset = viper.GetInt64(flag.Offset)
			c.recordCount = viper.GetInt(flag.RecordCount)
			c.recordNum = viper.GetInt(flag.RecordNum)
			c.showWarcHeader = viper.GetBool(flag.ShowWarcHeader)
			c.showProtocolHeader = viper.GetBool(flag.ShowProtocolHeader)
			c.showPayload = viper.GetBool(flag.ShowPayload)

			if (c.offset >= 0 || c.recordNum >= 0) && c.recordCount == 0 {
				c.recordCount = 1
			}
			if c.offset < 0 {
				c.offset = 0
			}

			c.filter = filter.NewFromViper()

			if !(c.showWarcHeader || c.showProtocolHeader || c.showPayload) {
				c.showWarcHeader = true
				c.showProtocolHeader = true
				c.showPayload = true
			}
			return runE(c)
		},
	}

	cmd.Flags().Int64P(flag.Offset, "o", -1, flag.OffsetHelp)
	cmd.Flags().IntP(flag.RecordNum, "n", -1, flag.RecordNumHelp)
	cmd.Flags().IntP(flag.RecordCount, "c", 0, flag.RecordCountHelp+" Defaults to show all records except if -o or -n option is set, then default is one.")
	cmd.Flags().BoolP(flag.ShowWarcHeader, "w", false, flag.ShowWarcHeaderHelp)
	cmd.Flags().BoolP(flag.ShowProtocolHeader, "p", false, flag.ShowProtocolHeaderHelp)
	cmd.Flags().BoolP(flag.ShowPayload, "P", false, flag.ShowPayloadHelp)
	cmd.Flags().StringArray(flag.RecordId, []string{}, flag.RecordIdHelp)
	cmd.Flags().StringSliceP(flag.RecordType, "t", []string{}, flag.RecordTypeHelp)
	cmd.Flags().StringP(flag.ResponseCode, "S", "", flag.ResponseCodeHelp)
	cmd.Flags().StringSliceP(flag.MimeType, "m", []string{}, flag.MimeTypeHelp)

	return cmd
}

func runE(c *conf) error {
	readFile(c, c.fileName)
	return nil
}

func readFile(c *conf, fileName string) {
	wf, err := gowarc.NewWarcFileReader(fileName, c.offset, gowarc.WithBufferTmpDir(viper.GetString(flag.TmpDir)))
	defer func() { _ = wf.Close() }()
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	num := 0
	count := 0

	for {
		wr, _, _, err := wf.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v, rec num: %v, Offset %v\n", err.Error(), strconv.Itoa(count), c.offset)
			break
		}

		if !c.filter.Accept(wr) {
			continue
		}

		// Find record number
		if c.recordNum > 0 && num < c.recordNum {
			num++
			continue
		}

		count++
		out := os.Stdout

		if c.showWarcHeader {
			// Write WARC record version
			_, err = fmt.Fprintf(out, "%v\r\n", wr.Version())
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			// Write WARC header
			_, err = wr.WarcHeader().Write(out)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			// Write separator
			_, err = out.Write([]byte("\r\n"))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}

		if c.showProtocolHeader {
			if b, ok := wr.Block().(gowarc.ProtocolHeaderBlock); ok {
				_, err = out.Write(b.ProtocolHeaderBytes())
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}
		}

		if c.showPayload {
			if pb, ok := wr.Block().(gowarc.PayloadBlock); ok {
				r, err := pb.PayloadBytes()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
				_, err = io.Copy(out, r)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			} else {
				r, err := wr.Block().RawBytes()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
				_, err = io.Copy(out, r)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}
		}

		// Write end of record separator
		_, err = out.Write([]byte("\r\n\r\n"))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		if c.recordCount > 0 && count >= c.recordCount {
			break
		}
	}
	_, _ = fmt.Fprintln(os.Stderr, "Count: ", count)
}
