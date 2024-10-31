package flag

import (
	"os"

	"github.com/nlnwa/gowarc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	TmpDir     = "tmpdir"
	TmpDirHelp = `directory to use for temporary files`
)

type WarcRecordOptionFlags struct {
}

func (f WarcRecordOptionFlags) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.String(TmpDir, os.TempDir(), TmpDirHelp)
}

func (f WarcRecordOptionFlags) TmpDir() string {
	return viper.GetString(TmpDir)
}

func (f WarcRecordOptionFlags) ToWarcRecordOptions() []gowarc.WarcRecordOption {
	return []gowarc.WarcRecordOption{
		gowarc.WithBufferTmpDir(f.TmpDir()),
	}
}
