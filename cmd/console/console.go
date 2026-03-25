package console

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/nationallibraryofnorway/warchaeology/v5/cmd/internal/flag"
	"github.com/nationallibraryofnorway/warchaeology/v5/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ConsoleFlags struct{}

func (f ConsoleFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String(flag.TempDir, os.TempDir(), flag.TempDirHelp)
	cmd.Flags().StringSlice(flag.Suffixes, []string{".warc", ".warc.gz"}, flag.SuffixesHelp)
}

func (f ConsoleFlags) TempDir() string {
	return viper.GetString(flag.TempDir)
}

func (f ConsoleFlags) Suffixes() []string {
	return viper.GetStringSlice(flag.Suffixes)
}

func (f ConsoleFlags) ToOptions(_ *cobra.Command, args []string) (*ui.Options, error) {
	if len(args) == 0 {
		return nil, errors.New("missing input directory")
	}

	dir := args[0]
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, err
	}

	fi, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}

	var files []string
	if !fi.IsDir() {
		files = append(files, filepath.Base(abs))
		abs = filepath.Dir(abs)
	}

	return &ui.Options{
		Dir:      abs,
		Files:    files,
		Suffixes: f.Suffixes(),
		TempDir:  f.TempDir(),
	}, nil
}

func NewCmdConsole() *cobra.Command {
	flags := ConsoleFlags{}

	cmd := &cobra.Command{
		Use:   "console DIR/FILE",
		Short: "A shell for working with WARC files",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.ToOptions(cmd, args)
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return ui.NewApp(opts).Run()
		},
		ValidArgsFunction: flag.SuffixCompletionFn,
	}

	flags.AddFlags(cmd)

	return cmd
}
