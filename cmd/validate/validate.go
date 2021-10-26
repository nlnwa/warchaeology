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

package validate

import (
	"context"
	"errors"
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

type conf struct {
	files          []string
	recursive      bool
	followSymlinks bool
	suffixes       []string
	concurrency    int
}

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:   "validate <files/dirs>",
		Short: "Validate warc files",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing file or directory name")
			}
			c.files = args
			return runE(c)
		},
	}

	cmd.Flags().BoolVarP(&c.recursive, flag.Recursive, "r", false, flag.RecursiveHelp)
	cmd.Flags().BoolVarP(&c.followSymlinks, flag.FollowSymlinks, "s", false, flag.FollowSymlinksHelp)
	cmd.Flags().StringSliceVar(&c.suffixes, flag.Suffixes, []string{".warc", ".warc.gz"}, flag.SuffixesHelp)
	cmd.Flags().IntVarP(&c.concurrency, flag.Concurrency, "c", int(float32(runtime.NumCPU())*float32(1.5)), flag.ConcurrencyHelp)

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

	fileWalker := filewalker.NewFromViper(c.files, validateFile)
	stats := filewalker.NewStats()
	return fileWalker.Walk(ctx, stats)
}

func validateFile(file string) filewalker.Result {
	result := filewalker.NewResult(file)

	wf, err := gowarc.NewWarcFileReader(file, 0)
	if err != nil {
		result.AddError(err)
		return result
	}
	defer func() { _ = wf.Close() }()

	for {
		wr, currentOffset, validation, err := wf.Next()
		if err == io.EOF {
			break
		}
		result.IncrRecords()
		result.IncrProcessed()
		if err != nil {
			result.AddError(fmt.Errorf("rec num: %d, offset: %d, cause: %w", result.Records(), currentOffset, err))
			break
		}

		err = wr.ValidateDigest(validation)
		if err != nil {
			result.AddError(fmt.Errorf("rec num: %d, offset: %d, cause: %w", result.Records(), currentOffset, err))
			break
		}

		if !validation.Valid() {
			result.AddError(fmt.Errorf("rec num: %d, offset: %d, record: %s, cause: %w", result.Records(), currentOffset, wr, validation))
		}
		_ = wr.Close()
	}
	return result
}
