package main

import (
	"runtime"

	"github.com/nlnwa/warchaeology/cmd"
	"github.com/nlnwa/warchaeology/internal/cmdversion"
)

// Overridden by ldflags
var (
	version = "dev"
	commit  = "none"
)

func main() {
	runtime.GOMAXPROCS(128)
	c := cmd.NewCommand()
	cmdversion.SetVersion(c.Name(), version, commit)
	c.SetVersionTemplate(`{{printf "%s\n" .Version}}`)
	c.Version = cmdversion.CmdVersion()
	_ = c.Execute()
}
