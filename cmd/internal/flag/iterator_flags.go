package flag

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Offset     = "offset"
	OffsetHelp = `record offset`

	RecordNum     = "num"
	RecordNumHelp = `print the n'th record. Only records that are not filtered out by other options are counted.`

	Limit     = "limit"
	LimitHelp = `The maximum number of records to show. Defaults to show all records.
If -o or -n option is set limit is set to 1.`
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
