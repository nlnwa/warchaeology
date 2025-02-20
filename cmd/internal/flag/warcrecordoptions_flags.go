package flag

import (
	"os"

	"github.com/nlnwa/gowarc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	TmpDir     = "tmp-dir"
	TmpDirHelp = `directory to use for temporary files`

	StrictValidation     = "strict"
	StrictValidationHelp = `sets the parser to fail on first validation error.`

	LenientValidation      = "lenient"
	LentientValidationHelp = `sets the parser to do as little validation as possible.`
)

type WarcRecordOptionFlags struct {
}

func (f WarcRecordOptionFlags) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.String(TmpDir, os.TempDir(), TmpDirHelp)
	flags.Bool(StrictValidation, false, StrictValidationHelp)
	flags.Bool(LenientValidation, false, LentientValidationHelp)
}

func (f WarcRecordOptionFlags) TmpDir() string {
	return viper.GetString(TmpDir)
}

func (f WarcRecordOptionFlags) StrictValidation() bool {
	return viper.GetBool(StrictValidation)
}

func (f WarcRecordOptionFlags) LenientValidation() bool {
	return viper.GetBool(LenientValidation)
}

func (f WarcRecordOptionFlags) ToWarcRecordOptions() []gowarc.WarcRecordOption {
	options := []gowarc.WarcRecordOption{
		gowarc.WithBufferTmpDir(f.TmpDir()),
	}
	if f.LenientValidation() {
		options = append(options, gowarc.WithNoValidation())
	}
	if f.StrictValidation() {
		options = append(options, gowarc.WithStrictValidation())
	}
	return options
}
