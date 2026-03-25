package flag

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ContinueOnError     = "continue-on-error"
	ContinueOnErrorHelp = `continue processing remaining files and directories after errors`
)

type ErrorFlags struct {
}

func (f ErrorFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(ContinueOnError, false, ContinueOnErrorHelp)
}

func (f ErrorFlags) ContinueOnError() bool {
	return viper.GetBool(ContinueOnError)
}
