package cmd

import (
	"os"

	"github.com/nlnwa/warchaeology/cmd/aart"
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
	cmd.AddCommand(aart.NewCommand())

	config.InitConfig(cmd)
	return cmd
}
