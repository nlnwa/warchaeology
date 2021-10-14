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

package dedup

import (
	"errors"
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/spf13/cobra"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type conf struct {
	recordTypes  gowarc.RecordType
	dirName      string
	indexDir     string
	writer       *gowarc.WarcFileWriter
	skipSymlinks bool
	index        *index
}

var recordTypes []string

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:   "dedup",
		Short: "Deduplicate WARC files",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, r := range recordTypes {
				switch strings.ToLower(r) {
				case "warcinfo":
					c.recordTypes = c.recordTypes | gowarc.Warcinfo
				case "request":
					c.recordTypes = c.recordTypes | gowarc.Request
				case "response":
					c.recordTypes = c.recordTypes | gowarc.Response
				case "metadata":
					c.recordTypes = c.recordTypes | gowarc.Metadata
				case "revisit":
					c.recordTypes = c.recordTypes | gowarc.Revisit
				case "resource":
					c.recordTypes = c.recordTypes | gowarc.Resource
				case "continuation":
					c.recordTypes = c.recordTypes | gowarc.Continuation
				case "conversion":
					c.recordTypes = c.recordTypes | gowarc.Conversion
				}
			}

			if len(args) == 0 {
				return errors.New("missing directory name")
			}
			c.dirName = args[0]
			n, err := cmd.Flags().GetBool("new-index")
			if err != nil {
				return err
			}
			c.index, err = newDb(c.indexDir, n)
			if err != nil {
				return err
			}
			defer c.index.close()

			return runE(c)
		},
	}

	cmd.Flags().StringSliceVarP(&recordTypes, "type", "t", []string{"response"}, "record types to"+
		" dedup. For more than one, repeat flag or comma separated list. Legal values: warcinfo,request,response,metadata,revisit,resource,continuation,conversion")
	cmd.Flags().BoolVarP(&c.skipSymlinks, "skip-symlinks", "s", false, "skip symlinks")
	cmd.Flags().Bool("new-index", false, "true to drop index on disk before dedup")
	cmd.Flags().StringVarP(&c.indexDir, "index-dir", "i", "/tmp/warc-dedup", "directory to store deduplication index")

	return cmd
}

func runE(c *conf) error {
	c.writer = gowarc.NewWarcFileWriter(
		gowarc.WithMaxConcurrentWriters(16),
		gowarc.WithCompression(true),
		//gowarc.WithMaxFileSize(internal.ConvertConf.MaxFileSize),
		//gowarc.WithFileNameGenerator(&namer),
		gowarc.WithFlush(false))

	//gowarc.WithMaxConcurrentWriters(internal.ConvertConf.ConcurrentWriters),
	//gowarc.WithCompression(internal.ConvertConf.Compress),
	//gowarc.WithMaxFileSize(internal.ConvertConf.MaxFileSize),
	//gowarc.WithFileNameGenerator(&namer),
	//gowarc.WithFlush(internal.ConvertConf.Flush))
	fmt.Println(c.writer)

	defer func(writer *gowarc.WarcFileWriter) {
		err := writer.Close()
		if err != nil {
			fmt.Printf("Error closing WARC writer: %v\n", err)
		}
	}(c.writer)

	stats := &stat{}
	pool := newWorkerpool(32)
	defer func() {
		pool.close()
		fmt.Fprintf(os.Stderr, "Count: %d, Dupes: %d\n", stats.count, stats.duplicates)
	}()
	result := make(chan *stat)
	go func() {
		for {
			s := <-result
			stats.count += s.count
			stats.duplicates += s.duplicates
		}
	}()

	walkDirs(c, c.dirName, pool, result)
	return nil
}

type stat struct {
	count, duplicates int
}

var fileCount int

func walkDirs(c *conf, dirName string, pool *workerpool, result chan<- *stat) {
	filepath.WalkDir(dirName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		if d.IsDir() {
			//fmt.Println(path)
		} else if strings.HasSuffix(path, ".warc") || strings.HasSuffix(path, ".warc.gz") {
			pool.submit(func() {
				fileCount++
				s := readFile(c, path, fileCount)
				result <- s
			})
		} else if !c.skipSymlinks && !d.IsDir() && !d.Type().IsRegular() {
			s, _ := filepath.EvalSymlinks(path)
			walkDirs(c, s, pool, result)
		}
		return nil
	})
	return
}

func readFile(c *conf, fileName string, filenum int) *stat {
	wf, err := gowarc.NewWarcFileReader(fileName, 0, gowarc.WithAddMissingDigest(true), gowarc.WithStrictValidation())
	defer func() { _ = wf.Close() }()
	if err != nil {
		panic(err)
	}

	count := 0
	duplicates := 0

	for {
		wr, currentOffset, _, err := wf.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v, rec num: %v, Offset %v\n", err.Error(), strconv.Itoa(count), currentOffset)
			break
		}

		if wr.Type()&c.recordTypes != 0 && wr.WarcHeader().Get(gowarc.ContentType) != "" {
			count++

			var digest string
			if wr.WarcHeader().Has(gowarc.WarcPayloadDigest) {
				digest = wr.WarcHeader().Get(gowarc.WarcPayloadDigest)
			} else if wr.WarcHeader().Has(gowarc.WarcBlockDigest) {
				digest = wr.WarcHeader().Get(gowarc.WarcBlockDigest)
			} else {
				fmt.Printf("Missing digest\n")
			}

			revisitRef := &ref{gowarc.RevisitRef{
				Profile:        gowarc.ProfileIdenticalPayloadDigestV1_1,
				TargetRecordId: wr.WarcHeader().Get(gowarc.WarcRecordID),
				TargetUri:      wr.WarcHeader().Get(gowarc.WarcTargetURI),
				TargetDate:     wr.WarcHeader().Get(gowarc.WarcDate),
			}}

			r, err := c.index.isRevisit(digest, revisitRef)
			if err != nil {
				fmt.Printf("Error getting revisit ref: %v\n", err)
			}
			if r != nil {
				if r.Profile == "" {
					panic(r)
				}
				duplicates++
				//fmt.Printf("DUP: %s => %s\n", revisitRef.TargetUri, r.TargetUri)
				revisit, err := wr.ToRevisitRecord(&r.RevisitRef)

				if err != nil {
					fmt.Printf("Error creating revisit record: %v\n", err)
				}
				if rr := c.writer.Write(revisit); rr != nil && rr[0].Err != nil {
					fmt.Printf("Offset: %d\n", currentOffset)
					wr.WarcHeader().Write(os.Stdout)
					panic(rr[0].Err)
				}
			} else {
				if rr := c.writer.Write(wr); rr != nil && rr[0].Err != nil {
					panic(rr[0].Err)
				}
			}
		} else {
			if rr := c.writer.Write(wr); rr != nil && rr[0].Err != nil {
				panic(rr[0].Err)
			}
		}
		wr.Close()
	}

	fmt.Fprintf(os.Stderr, " %04d > %s Count: %d, Dupes: %d\n", filenum, fileName, count, duplicates)
	return &stat{
		count:      count,
		duplicates: duplicates,
	}
}
