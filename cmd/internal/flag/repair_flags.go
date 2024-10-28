package flag

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Repair     = "repair"
	RepairHelp = `try to fix errors in records`
)

type RepairFlags struct {
}

func (u RepairFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP(Repair, "R", false, RepairHelp)
}

func (u RepairFlags) Repair() bool {
	return viper.GetBool(Repair)
}
