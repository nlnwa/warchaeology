package main

import (
	"github.com/nlnwa/warchaeology/cmd"
	cmdversion "github.com/nlnwa/warchaeology/internal/version"
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

	_ = cmd.Execute()
}
