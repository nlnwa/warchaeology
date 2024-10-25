package flag

import (
	"github.com/nlnwa/warchaeology/v3/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	MinFreeDisk     = "min-free-disk"
	MinFreeDiskHelp = `minimum free space on disk to allow WARC writing`

	Repair     = "repair"
	RepairHelp = `try to fix errors in records`
)

type UtilFlags struct {
}

func (u UtilFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String(MinFreeDisk, "256MB", MinFreeDiskHelp)
	cmd.Flags().BoolP(Repair, "R", false, RepairHelp)
}

func (u UtilFlags) MinFreeDisk() int64 {
	minFreeDisk := viper.GetString(MinFreeDisk)
	return util.ParseSizeInBytes(minFreeDisk)
}

func (u UtilFlags) Repair() bool {
	return viper.GetBool(Repair)
}