package convert

import (
	"github.com/nlnwa/warchaeology/cmd/convert/arc"
	"github.com/nlnwa/warchaeology/cmd/convert/nedlib"
	"github.com/nlnwa/warchaeology/cmd/convert/warc"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert web archives to warc files. Use subcommands for the supported formats",
		Long:  ``,
	}

	// Subcommands
	cmd.AddCommand(nedlib.NewCommand())
	cmd.AddCommand(arc.NewCommand())
	cmd.AddCommand(warc.NewCommand())

	return cmd
}
