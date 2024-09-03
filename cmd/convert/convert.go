package convert

import (
	"github.com/nlnwa/warchaeology/v3/cmd/convert/arc"
	"github.com/nlnwa/warchaeology/v3/cmd/convert/nedlib"
	"github.com/nlnwa/warchaeology/v3/cmd/convert/warc"
	"github.com/spf13/cobra"
)

func NewCmdConvert() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert web archives to warc files. Use subcommands for the supported formats",
		Long:  ``,
	}

	// Subcommands
	cmd.AddCommand(nedlib.NewCmdConvertNedlib())
	cmd.AddCommand(arc.NewCommand())
	cmd.AddCommand(warc.NewCommand())

	return cmd
}
