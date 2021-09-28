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
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/cmd/convert/internal"
	"github.com/spf13/cobra"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type conf struct {
	dir string
}

const dateFormat = "2006-1-2"

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:   "nedlib <dir>",
		Short: "Convert directory with files harvested with Nedlib into warc files",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing directory name")
			}
			c.dir = args[0]
			return runE(c)
		},
	}

	return cmd
}

func runE(c *conf) error {
	if f, err := os.Lstat(internal.ConvertConf.OutDir); err != nil {
		return fmt.Errorf("could not write to output directory '%s': %w", internal.ConvertConf.OutDir, err.(*os.PathError).Err)
	} else if !f.IsDir() {
		return fmt.Errorf("could not write to output directory: '%s' is not a directory", internal.ConvertConf.OutDir)
	}

	namer := &gowarc.PatternNameGenerator{
		Directory: internal.ConvertConf.OutDir,
		Prefix:    internal.ConvertConf.FilePrefix,
	}
	writer := gowarc.NewWarcFileWriter(
		gowarc.WithMaxConcurrentWriters(internal.ConvertConf.ConcurrentWriters),
		gowarc.WithCompression(internal.ConvertConf.Compress),
		gowarc.WithMaxFileSize(internal.ConvertConf.MaxFileSize),
		gowarc.WithFileNameGenerator(namer),
		gowarc.WithFlush(internal.ConvertConf.Flush))
	fmt.Println(writer)

	defer func(writer *gowarc.WarcFileWriter) {
		err := writer.Close()
		if err != nil {
			fmt.Printf("Error closing WARC writer: %v\n", err)
		}
	}(writer)

	wp := NewWorkerPool(32)
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go wp.Run(ctx)

	count := 0
	errors := 0

	go func() {
		err := filepath.WalkDir(c.dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("%s, %v, %v", path, d, err)
				return err
			}
			if d.IsDir() {
				fmt.Printf("\nWorking on dir: '%s'\n", path)
			} else {
				fmt.Print(".")
				if strings.HasSuffix(path, ".meta") {
					count++

					j := Job{
						ID:           count,
						ExecFn:       writeRecord,
						writer:       writer,
						config:       c,
						metaFileName: path,
					}
					wp.jobs <- j
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			errors++
			//return err
		}
		fmt.Println()

		close(wp.jobs)
	}()

	handleResults(wp)
	_, _ = fmt.Fprintf(os.Stderr, "\nCount: %d, Errors: %d\n", count, errors)
	return nil
}

func walkDir() {

}

func handleResults(wp WorkerPool) {
	for {
		select {
		case _, ok := <-wp.Results():
			if !ok {
				continue
			}

		case <-wp.Done:
			return
		default:
		}
	}
}

func writeRecord(ctx context.Context, writer *gowarc.WarcFileWriter, config *conf, metaFileName string) (string, error) {
	f, err := os.Open(metaFileName)
	if err != nil {
		return "", err
	}

	response, err := http.ReadResponse(bufio.NewReader(io.MultiReader(f, bytes.NewReader([]byte{'\r', '\n'}))), nil)
	f.Close()
	if err != nil {
		return "", err
	}

	rb := gowarc.NewRecordBuilder(gowarc.Response,
		gowarc.WithAddMissingDigest(true),
		gowarc.WithFixDigest(true),
		gowarc.WithFixContentLength(true),
		gowarc.WithVersion(internal.ConvertConf.WarcVersion))

	rb.AddWarcHeader(gowarc.ContentType, "application/http;msgtype=response")

	header := response.Header
	dateString := header.Get("Date")
	if dateString != "" {
		t, err := time.Parse(time.RFC1123, dateString)
		if err != nil {
			return "", err
		}
		rb.AddWarcHeaderTime(gowarc.WarcDate, t)
	} else {
		rb.AddWarcHeaderTime(gowarc.WarcDate, internal.ConvertConf.DefaultTime)
	}

	for i, _ := range header {
		if strings.HasPrefix(i, "Arc") {
			if i == "Arc-Url" {
				rb.AddWarcHeader(gowarc.WarcTargetURI, header.Get(i))
			}
			header.Del(i)
		}
	}
	rb.WriteString(response.Proto + " " + response.Status + "\n")
	header.Write(rb)

	rb.WriteString("\r\n")

	p, err := os.Open(strings.TrimSuffix(metaFileName, ".meta"))
	defer p.Close()
	if err != nil {
		return "", err
	}

	rb.ReadFrom(p)

	wr, _, err := rb.Build()
	defer wr.Close()

	resp := writer.Write(wr)
	if resp[0].Err != nil {
		fmt.Printf("%s - %v", wr, resp[0].Err)
	}

	return fmt.Sprintf("%s - %v", wr, resp[0].Err), nil
}
