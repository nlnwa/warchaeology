package flag

import (
	"github.com/nlnwa/warchaeology/v4/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	MinDiskFree     = "min-disk-free"
	MinDiskFreeHelp = `minimum free space on disk to allow WARC writing`
)

type UtilFlags struct {
}

func (u UtilFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String(MinDiskFree, "1GB", MinDiskFreeHelp)
}

func (u UtilFlags) MinFreeDisk() int64 {
	minFreeDisk := viper.GetString(MinDiskFree)
	return util.ParseSizeInBytes(minFreeDisk)
}
