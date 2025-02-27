package flag

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Offset     = "offset"
	OffsetHelp = `start processing from this byte offset in file. Defaults to 0.`

	RecordNum     = "nth"
	RecordNumHelp = `only process the n'th record. Only records that are not filtered out by other options are counted.`

	Limit     = "limit"
	LimitHelp = `limit the number of records to process. If the -n option is specified the limit is ignored.`

	Force     = "force"
	ForceHelp = `force the record iterator to continue regardless of errors.`
)

type WarcIteratorFlags struct {
}

func NewWarcIteratorFlags() WarcIteratorFlags {
	return WarcIteratorFlags{}
}

func (f WarcIteratorFlags) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.Int64P(Offset, "o", 0, OffsetHelp)
	flags.IntP(RecordNum, "n", 0, RecordNumHelp)
	flags.IntP(Limit, "l", 0, LimitHelp)
	flags.BoolP(Force, "f", false, ForceHelp)
}

func (f WarcIteratorFlags) Offset() int64 {
	return viper.GetInt64(Offset)
}

func (f WarcIteratorFlags) RecordNum() int {
	return viper.GetInt(RecordNum)
}

func (f WarcIteratorFlags) Limit() int {
	return viper.GetInt(Limit)
}

func (f WarcIteratorFlags) Force() bool {
	return viper.GetBool(Force)
}
