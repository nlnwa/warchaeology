/*
 * Copyright © 2019 National Library of Norway
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"os"

	"github.com/nlnwa/warchaeology/cmd/cat"
	"github.com/nlnwa/warchaeology/cmd/console"
	"github.com/nlnwa/warchaeology/cmd/convert"
	"github.com/nlnwa/warchaeology/cmd/dedup"
	"github.com/nlnwa/warchaeology/cmd/ls"
	"github.com/nlnwa/warchaeology/cmd/validate"
	"github.com/nlnwa/warchaeology/internal/config"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/cobra"
)

// NewCommand returns a new cobra.Command implementing the root command for warc
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "warc",
		Short: "A tool for handling warc files",
		Long:  ``,
	}

	// Flags
	cmd.PersistentFlags().StringP(flag.LogFileName, "L", "", flag.LogFileNameHelp)
	cmd.PersistentFlags().StringSlice(flag.LogFile, []string{"info", "error", "summary"}, flag.LogFileHelp)
	cmd.PersistentFlags().StringSlice(flag.LogConsole, []string{"progress", "summary"}, flag.LogConsoleHelp)
	cmd.PersistentFlags().String(flag.TmpDir, os.TempDir(), flag.TmpDirHelp)
	cmd.PersistentFlags().String(flag.BufferMaxMem, "1MB", flag.BufferMaxMemHelp)
	_ = cmd.RegisterFlagCompletionFunc(flag.LogFile, flag.SliceCompletion{
		"info\tShow stats for each file",
		"error\tPrint errors",
		"summary\tCreate a summary after completion",
	}.CompletionFn)
	_ = cmd.RegisterFlagCompletionFunc(flag.LogConsole, flag.SliceCompletion{
		"info\tShow stats for each file",
		"error\tPrint errors",
		"summary\tCreate a summary after completion",
		"progress\tShow progress while running",
	}.CompletionFn)

	// Subcommands
	cmd.AddCommand(ls.NewCommand())
	cmd.AddCommand(cat.NewCommand())
	cmd.AddCommand(validate.NewCommand())
	cmd.AddCommand(console.NewCommand())
	cmd.AddCommand(convert.NewCommand())
	cmd.AddCommand(dedup.NewCommand())
	cmd.AddCommand(completionCmd)

	config.InitConfig(cmd)
	return cmd
}
