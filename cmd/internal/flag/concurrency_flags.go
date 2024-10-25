package flag

import (
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Concurrency     = "concurrency"
	ConcurrencyHelp = `number of input files to process simultaneously.`
)

type ConcurrencyFlags struct{}

func (f ConcurrencyFlags) AddFlags(cmd *cobra.Command) {
	defaultConcurrency := int(float32(runtime.NumCPU()) * float32(1.5))

	flags := cmd.Flags()
	flags.IntP(Concurrency, "c", defaultConcurrency, ConcurrencyHelp)
}

func (f ConcurrencyFlags) Concurrency() int {
	return viper.GetInt(Concurrency)
}
