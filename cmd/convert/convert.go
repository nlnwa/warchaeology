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

package convert

import (
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/cmd/convert/arc"
	"github.com/nlnwa/warchaeology/cmd/convert/internal"
	"github.com/nlnwa/warchaeology/cmd/convert/nedlib"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

const dateFormat = "2006-1-2"

var (
	defaultTimeString string
	warcVersion       string
)

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert web archives to warc files. Use subcommands for the supported formats",
		Long:  ``,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}

			if t, err := time.Parse(dateFormat, defaultTimeString); err != nil {
				return err
			} else {
				internal.ConvertConf.DefaultTime = t.Add(12 * time.Hour)
			}

			switch warcVersion {
			case "1.0":
				internal.ConvertConf.WarcVersion = gowarc.V1_0
			case "1.1":
				internal.ConvertConf.WarcVersion = gowarc.V1_1
			default:
				return fmt.Errorf("unknown WARC version: %s", warcVersion)
			}

			var err error
			if internal.ConvertConf.OutDir, err = filepath.Abs(internal.ConvertConf.OutDir); err != nil {
				return err
			}
			if internal.ConvertConf.OutDir, err = filepath.EvalSymlinks(internal.ConvertConf.OutDir); err != nil {
				return err
			}
			if f, err := os.Lstat(internal.ConvertConf.OutDir); err != nil {
				return fmt.Errorf("could not write to output directory '%s': %w", internal.ConvertConf.OutDir, err.(*os.PathError).Err)
			} else if !f.IsDir() {
				return fmt.Errorf("could not write to output directory: '%s' is not a directory", internal.ConvertConf.OutDir)
			}

			return nil
		},
	}

	cmd.PersistentFlags().IntVarP(&internal.ConvertConf.ConcurrentWriters, "concurrent-writers", "c", 1, "maximum concurrent WARC writers")
	cmd.PersistentFlags().Int64VarP(&internal.ConvertConf.MaxFileSize, "file-size", "s", 1024*1024*1024, "The maximum size for WARC files")
	cmd.PersistentFlags().BoolVarP(&internal.ConvertConf.Compress, "compress", "z", false, "use gzip compression for WARC files")
	cmd.PersistentFlags().StringVarP(&internal.ConvertConf.FilePrefix, "prefix", "p", "", "filename prefix for WARC files")
	cmd.PersistentFlags().StringVarP(&defaultTimeString, "time", "t", time.Now().Format(dateFormat), "fetch date to use for records missing date metadata. Fetchtime is set to 12:00 UTC for the date")
	cmd.PersistentFlags().StringVarP(&internal.ConvertConf.OutDir, "warc-dir", "w", ".", "output directory for generated warc files. Directory must exist.")
	cmd.PersistentFlags().StringVar(&warcVersion, "warc-version", "1.1", "the WARC version to use for created files")
	cmd.PersistentFlags().BoolVar(&internal.ConvertConf.Flush, "flush", false, "if true, sync WARC file to disk after writing each record")

	// Subcommands
	cmd.AddCommand(nedlib.NewCommand())
	cmd.AddCommand(arc.NewCommand())

	return cmd
}
