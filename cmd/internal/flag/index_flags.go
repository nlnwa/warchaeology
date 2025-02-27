package flag

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nlnwa/warchaeology/v3/internal/index"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	NewIndex     = "new-index"
	NewIndexHelp = `true to start from a fresh index, deleting eventual index from last run`

	KeepIndex     = "keep-index"
	KeepIndexHelp = `true to keep index on disk so that the next run will continue where the previous run left off`

	IndexDir     = "index-dir"
	IndexDirHelp = `directory to store indexes`
)

type IndexFlags struct {
	name string
}

func (f *IndexFlags) AddFlags(cmd *cobra.Command, options ...func(*IndexFlags)) {
	f.name = cmd.Name()

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(fmt.Errorf("failed to get user cache dir: %v", err))
	}

	cacheDir = filepath.Join(cacheDir, f.name)
	if f.name == "" {
		cacheDir = filepath.Join(cacheDir, "index")
	}

	flags := cmd.Flags()
	flags.BoolP(KeepIndex, "k", false, KeepIndexHelp)
	flags.BoolP(NewIndex, "K", false, NewIndexHelp)
	flags.String(IndexDir, cacheDir, IndexDirHelp)

	if err := cmd.MarkFlagDirname(IndexDir); err != nil {
		panic(err)
	}
}

func (f *IndexFlags) IndexDir() string {
	return viper.GetString(IndexDir)
}

func (f *IndexFlags) KeepIndex() bool {
	return viper.GetBool(KeepIndex)
}

func (f *IndexFlags) NewIndex() bool {
	return viper.GetBool(NewIndex)
}

func (f *IndexFlags) ToDigestIndex() (*index.DigestIndex, error) {
	return index.NewDigestIndex(f.IndexDir(), f.name, f.KeepIndex(), f.NewIndex())
}

func (f *IndexFlags) ToFileIndex() (*index.FileIndex, error) {
	return index.NewFileIndex(f.IndexDir(), f.name, f.KeepIndex(), f.NewIndex())
}
