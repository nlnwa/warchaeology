package convert

import (
	"github.com/nlnwa/warchaeology/v4/cmd/convert/arc"
	"github.com/nlnwa/warchaeology/v4/cmd/convert/nedlib"
	"github.com/nlnwa/warchaeology/v4/cmd/convert/warc"
	"github.com/spf13/cobra"
)

func NewCmdConvert() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert web archive files to WARC files. Use subcommands for the supported formats",
		Long:  ``,
	}

	// Subcommands
	cmd.AddCommand(nedlib.NewCmdConvertNedlib())
	cmd.AddCommand(arc.NewCmdConvertArc())
	cmd.AddCommand(warc.NewCmdConvertWarc())

	return cmd
}
