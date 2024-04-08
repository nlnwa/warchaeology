package listRecord

import (
	"fmt"

	listRecordImplementation "github.com/nlnwa/warchaeology/internal/listrecord"
	"github.com/spf13/cobra"
)

const (
	warcFiles = "warc-file"
)

func NewCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list-records",
		Short: "List WARC file records",
		Args:  cobra.NoArgs,
		RunE:  parseArgumentsAndCallListRecordsAndPrint,
	}

	command.Flags().String(warcFiles, "", "file to list records from")
	if markWarcFilesRequiredError := command.MarkFlagRequired(warcFiles); markWarcFilesRequiredError != nil {
		panic(fmt.Sprintf("failed to mark flag %s as required, original error: '%s'", warcFiles, markWarcFilesRequiredError))
	}

	return command
}

func parseArgumentsAndCallListRecordsAndPrint(command *cobra.Command, args []string) error {
	warcFile, err := command.Flags().GetString(warcFiles)
	if err != nil {
		return fmt.Errorf("getting warc-file path failed, original error: '%w'", err)
	}
	warcRecords, err := listRecordImplementation.ListRecords(warcFile)
	if err != nil {
		return fmt.Errorf("listing records failed, original error: '%w'", err)
	}

	fmt.Printf("Records in path '%s':\n", warcFile)
	for _, warcRecord := range warcRecords {
		if !warcRecord.Validation.Valid() {
			fmt.Printf("Record with offset '%v' is invalid\n", warcRecord.Offset)
		}
		println(warcRecord.Record.String())
	}

	return nil
}
