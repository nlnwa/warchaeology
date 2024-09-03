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

type ConcurrencyFlags struct {
	defaultConcurrency int
}

func WithDefaultConcurrency(defaultConcurrency int) func(*ConcurrencyFlags) {
	return func(f *ConcurrencyFlags) {
		f.defaultConcurrency = defaultConcurrency
	}
}

func (f ConcurrencyFlags) AddFlags(cmd *cobra.Command, opts ...func(*ConcurrencyFlags)) {
	for _, opt := range opts {
		opt(&f)
	}
	if f.defaultConcurrency == 0 {
		f.defaultConcurrency = int(float32(runtime.NumCPU()) * float32(1.5))
	}

	flags := cmd.Flags()
	flags.IntP(Concurrency, "c", f.defaultConcurrency, ConcurrencyHelp)
}

func (f ConcurrencyFlags) Concurrency() int {
	return viper.GetInt(Concurrency)
}
