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
	"errors"
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/arcreader"
	"github.com/nlnwa/warchaeology/cmd/convert/internal"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path"
	"strings"
)

type conf struct {
	fileName string
}

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:   "arc <file>",
		Short: "Convert arc file into warc file",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing directory name")
			}
			c.fileName = args[0]
			return runE(c)
		},
	}

	return cmd
}

func runE(c *conf) error {
	f := path.Base(c.fileName)
	f = strings.TrimSuffix(f, ".gz")
	f = strings.ReplaceAll(f, ".arc", ".warc")
	namer := arc2warcNamer(f)

	writer := gowarc.NewWarcFileWriter(
		gowarc.WithMaxConcurrentWriters(internal.ConvertConf.ConcurrentWriters),
		gowarc.WithCompression(internal.ConvertConf.Compress),
		gowarc.WithMaxFileSize(internal.ConvertConf.MaxFileSize),
		gowarc.WithFileNameGenerator(&namer),
		gowarc.WithFlush(internal.ConvertConf.Flush))
	fmt.Println(writer)

	defer func(writer *gowarc.WarcFileWriter) {
		err := writer.Close()
		if err != nil {
			fmt.Printf("Error closing WARC writer: %v\n", err)
		}
	}(writer)

	count := 0
	errors := 0

	a, err := arcreader.NewArcFileReader(c.fileName, 0, gowarc.WithStrictValidation(), gowarc.WithVersion(internal.ConvertConf.WarcVersion))
	if err != nil {
		return err
	}
	defer a.Close()

	for {
		wr, _, _, err := a.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		resp := writer.Write(wr)
		if resp[0].Err != nil {
			fmt.Printf("Err: %v\n", resp[0].Err)
			errors++
		}
		count++
		wr.Close()
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nCount: %d, Errors: %d\n", count, errors)
	return nil
}

type arc2warcNamer string

func (a arc2warcNamer) NewWarcfileName() (string, string) {
	return internal.ConvertConf.OutDir, string(a)
}
