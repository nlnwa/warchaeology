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

package create

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/nlnwa/gowarc"
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
	dir               string
	fileName          string
	compress          bool
	concurrentWriters int
	maxFileSize       int64
	filePrefix        string
	defaultTimeString string
	defaultTime       time.Time
	outDir            string
}

const dateFormat = "2006-1-2"

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:   "create <dir>",
		Short: "Create warc files",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing directory name")
			}
			c.dir = args[0]
			if t, err := time.Parse(dateFormat, c.defaultTimeString); err != nil {
				return err
			} else {
				c.defaultTime = t.Add(12 * time.Hour)
			}
			return runE(c)
		},
	}

	cmd.Flags().IntVarP(&c.concurrentWriters, "concurrent-writers", "c", 1, "maximum concurrent WARC writers")
	cmd.Flags().Int64VarP(&c.maxFileSize, "file-size", "s", 1024*1024*1024, "The maximum size for WARC files")
	cmd.Flags().BoolVarP(&c.compress, "compress", "z", false, "use gzip compression for WARC files")
	cmd.Flags().StringVarP(&c.filePrefix, "prefix", "p", "", "filename prefix for WARC files")
	cmd.Flags().StringVarP(&c.defaultTimeString, "time", "t", time.Now().Format(dateFormat), "fetch date to use for records missing date metadata. Fetchtime is set to 12:00 UTC for the date")
	cmd.Flags().StringVarP(&c.outDir, "warc-dir", "w", ".", "output directory for generated warc files. Directory must exist.")

	return cmd
}

func runE(c *conf) error {
	if f, err := os.Lstat(c.outDir); err != nil {
		return fmt.Errorf("could not write to output directory '%s': %w", c.outDir, err.(*os.PathError).Err)
	} else if !f.IsDir() {
		return fmt.Errorf("could not write to output directory: '%s' is not a directory", c.outDir)
	}

	namer := &gowarc.PatternNameGenerator{
		Directory: c.outDir,
		Prefix:    c.filePrefix,
	}
	writer := gowarc.NewWarcFileWriter(
		gowarc.WithMaxConcurrentWriters(c.concurrentWriters),
		gowarc.WithCompression(c.compress),
		gowarc.WithMaxFileSize(c.maxFileSize),
		gowarc.WithFileNameGenerator(namer))
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

	rb := gowarc.NewRecordBuilder(gowarc.Response, gowarc.WithAddMissingDigest(true), gowarc.WithFixDigest(true), gowarc.WithFixContentLength(true))
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
		rb.AddWarcHeaderTime(gowarc.WarcDate, config.defaultTime)
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
