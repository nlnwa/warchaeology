package cmdversion

import "fmt"

var Name string

// Version contains the version of this app
var Version string

var Commit string

func SetVersion(name, version, commit string) {
	Name = name
	Version = version
	Commit = commit
}

func CmdVersion() string {
	return fmt.Sprintf("%s %s (commit: %s)", Name, Version, Commit)
}

func SoftwareVersion() string {
	return fmt.Sprintf("warchaeology/%s", Version)
}
