package cat

import (
	"errors"

	"github.com/nlnwa/warchaeology/internal/filter"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cat",
		Short: "Concatenate and print warc files",
		Long:  ``,
		Example: `# Print all content from a WARC file
warc cat file1.warc.gz

# Pipe payload from record #4 into the image viewer feh
warc cat -n4 -P file1.warc.gz | feh -`,
		RunE: parseArgumentsAndCallCat,
	}

	cmd.Flags().Int64P(flag.Offset, "o", -1, flag.OffsetHelp)
	cmd.Flags().IntP(flag.RecordNum, "n", -1, flag.RecordNumHelp)
	cmd.Flags().IntP(flag.RecordCount, "c", 0, flag.RecordCountHelp+" Defaults to show all records except if -o or -n option is set, then default is one.")
	cmd.Flags().BoolP(flag.ShowWarcHeader, "w", false, flag.ShowWarcHeaderHelp)
	cmd.Flags().BoolP(flag.ShowProtocolHeader, "p", false, flag.ShowProtocolHeaderHelp)
	cmd.Flags().BoolP(flag.ShowPayload, "P", false, flag.ShowPayloadHelp)
	cmd.Flags().StringArray(flag.RecordId, []string{}, flag.RecordIdHelp)
	cmd.Flags().StringSliceP(flag.RecordType, "t", []string{}, flag.RecordTypeHelp)
	cmd.Flags().StringP(flag.ResponseCode, "S", "", flag.ResponseCodeHelp)
	cmd.Flags().StringSliceP(flag.MimeType, "m", []string{}, flag.MimeTypeHelp)

	return cmd
}

func parseArgumentsAndCallCat(cmd *cobra.Command, args []string) error {
	catConfig := &config{}
	if len(args) == 0 {
		return errors.New("missing file name")
	}
	catConfig.fileName = args[0]
	catConfig.offset = viper.GetInt64(flag.Offset)
	catConfig.recordCount = viper.GetInt(flag.RecordCount)
	catConfig.recordNum = viper.GetInt(flag.RecordNum)
	catConfig.showWarcHeader = viper.GetBool(flag.ShowWarcHeader)
	catConfig.showProtocolHeader = viper.GetBool(flag.ShowProtocolHeader)
	catConfig.showPayload = viper.GetBool(flag.ShowPayload)

	if (catConfig.offset >= 0 || catConfig.recordNum >= 0) && catConfig.recordCount == 0 {
		catConfig.recordCount = 1
	}
	if catConfig.offset < 0 {
		catConfig.offset = 0
	}

	catConfig.filter = filter.NewFromViper()

	if !(catConfig.showWarcHeader || catConfig.showProtocolHeader || catConfig.showPayload) {
		catConfig.showWarcHeader = true
		catConfig.showProtocolHeader = true
		catConfig.showPayload = true
	}

	// Silence usage to avoid printing usage when an error occurs
	cmd.SilenceUsage = true

	return runCat(catConfig)
}
