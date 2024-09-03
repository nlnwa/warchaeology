package flag

import (
	"fmt"

	"github.com/nlnwa/warchaeology/v3/internal/filewalker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Recursive     = "recursive"
	RecursiveHelp = "walk directories recursively"

	FollowSymlinks     = "symlinks"
	FollowSymlinksHelp = `follow symlinks`

	Suffixes     = "suffixes"
	SuffixesHelp = `filter files by suffix`
)

type FileWalkerFlags struct {
	SrcFileListFlags SrcFileListFlags
	suffixes         []string
}

func WithDefaultSuffixes(suffixes []string) func(*FileWalkerFlags) {
	return func(f *FileWalkerFlags) {
		f.suffixes = suffixes
	}
}

func (f FileWalkerFlags) AddFlags(cmd *cobra.Command, options ...func(*FileWalkerFlags)) {
	for _, option := range options {
		option(&f)
	}
	if len(f.suffixes) == 0 {
		f.suffixes = []string{".warc", ".warc.gz"}
	}

	flags := cmd.Flags()
	flags.BoolP(Recursive, "r", false, RecursiveHelp)
	flags.BoolP(FollowSymlinks, "s", false, FollowSymlinksHelp)
	flags.StringSlice(Suffixes, f.suffixes, SuffixesHelp)

	f.SrcFileListFlags.AddFlags(cmd)
}

func (f FileWalkerFlags) ToFileWalker() (*filewalker.FileWalker, error) {
	fs, err := f.SrcFileListFlags.ToFs()
	if err != nil {
		return nil, fmt.Errorf("failed to create file system: %w", err)
	}

	return filewalker.New(
		filewalker.WithFs(fs),
		filewalker.WithRecursive(viper.GetBool(Recursive)),
		filewalker.WithFollowSymlinks(viper.GetBool(FollowSymlinks)),
		filewalker.WithSuffixes(viper.GetStringSlice(Suffixes)),
	), nil
}
