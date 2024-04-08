package cmd

import (
	"github.com/nlnwa/warchaeology/cmd/aart"
	"github.com/nlnwa/warchaeology/cmd/cat"
	"github.com/nlnwa/warchaeology/cmd/console"
	"github.com/nlnwa/warchaeology/cmd/convert"
	"github.com/nlnwa/warchaeology/cmd/dedup"
	listRecord "github.com/nlnwa/warchaeology/cmd/listrecords"
	"github.com/nlnwa/warchaeology/cmd/validate"
	"github.com/spf13/cobra"
)

// NewCommand returns a new cobra.Command implementing the root command for warc
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "warc",
		Short: "A tool for handling warc files",
		Long:  ``,
	}

	// Subcommands
	cmd.AddCommand(listRecord.NewCommand())
	cmd.AddCommand(cat.NewCommand())
	cmd.AddCommand(validate.NewCommand())
	cmd.AddCommand(console.NewCommand())
	cmd.AddCommand(convert.NewCommand())
	cmd.AddCommand(dedup.NewCommand())
	cmd.AddCommand(completionCmd)
	cmd.AddCommand(aart.NewCommand())

	return cmd
}
