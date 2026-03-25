package flag

import (
	"github.com/nlnwa/gowarc/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Repair     = "repair"
	RepairHelp = `attempt to repair malformed records when possible`
)

type RepairFlags struct {
}

func (r RepairFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP(Repair, "R", false, RepairHelp)
}

func (r RepairFlags) Repair() bool {
	return viper.GetBool(Repair)
}

func (r RepairFlags) ToWarcRecordOptions() []gowarc.WarcRecordOption {
	if !r.Repair() {
		return nil
	}
	return []gowarc.WarcRecordOption{
		gowarc.WithSyntaxErrorPolicy(gowarc.ErrWarn),
		gowarc.WithSpecViolationPolicy(gowarc.ErrWarn),
		gowarc.WithAddMissingDigest(true),
		gowarc.WithFixSyntaxErrors(true),
		gowarc.WithFixDigest(true),
		gowarc.WithAddMissingContentLength(true),
		gowarc.WithAddMissingRecordId(true),
		gowarc.WithFixContentLength(true),
		gowarc.WithFixWarcFieldsBlockErrors(true),
	}
}
