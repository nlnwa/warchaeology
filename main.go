package main

import (
	"os"

	"github.com/nlnwa/warchaeology/v3/cmd"
	cmdversion "github.com/nlnwa/warchaeology/v3/internal/version"
)

var (
	version = "dirty"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd := cmd.NewWarcCommand()

	cmdversion.Set(version, commit, date)
	cmd.Version = cmdversion.Version.GitVersion
	cmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
