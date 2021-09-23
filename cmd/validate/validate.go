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
	"errors"
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strconv"
)

type conf struct {
	fileName string
}

func NewCommand() *cobra.Command {
	c := &conf{}
	var cmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate warc files",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing file name")
			}
			c.fileName = args[0]
			return runE(c)
		},
	}

	return cmd
}

func runE(c *conf) error {
	wf, err := gowarc.NewWarcFileReader(c.fileName, 0)
	if err != nil {
		return err
	}
	defer func() { _ = wf.Close() }()

	count := 0
	errorRecords := 0

	for {
		wr, currentOffset, validation, err := wf.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v, rec num: %v, Offset %v\n", err.Error(), strconv.Itoa(count), currentOffset)
			break
		}
		count++

		err = wr.ValidateDigest(validation)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v, rec num: %v, Offset %v\n", err.Error(), strconv.Itoa(count), currentOffset)
			break
		}

		if !validation.Valid() {
			errorRecords++
			fmt.Printf("Offset: %d, %s:\n%v\n", currentOffset, wr, validation)
		}
		_ = wr.Close()
	}
	_, _ = fmt.Fprintf(os.Stderr, "Count: %d, Errors: %d\n", count, errorRecords)
	return nil
}
