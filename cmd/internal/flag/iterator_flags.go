package flag

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Offset     = "offset"
	OffsetHelp = `start processing at this byte offset in the input file (default: 0)`

	RecordNum     = "nth"
	RecordNumHelp = `process only the n-th record after filtering`

	Limit     = "limit"
	LimitHelp = `maximum number of records to process; ignored when --nth is set`

	Force     = "force"
	ForceHelp = `continue iterating even when record read errors occur`
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
