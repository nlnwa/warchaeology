package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var Version Info

type Info struct {
	GitVersion string `json:"gitVersion" yaml:"gitVersion"`
	GitCommit  string `json:"gitCommit"  yaml:"gitCommit"`
	BuildDate  string `json:"buildDate"  yaml:"buildDate"`
	GoVersion  string `json:"goVersion"  yaml:"goVersion"`
	Compiler   string `json:"compiler"   yaml:"compiler"`
	Platform   string `json:"platform"   yaml:"platform"`
}

func (v Info) String() string {
	return fmt.Sprintf(`Git Version: %s
Git commit: %s
Build date: %s
Go version: %s
Compiler: %s
Platform: %s
`, v.GitVersion, v.GitCommit, v.BuildDate, runtime.Version(), runtime.Compiler, fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
}

func Set(version, commit, date string) {
	Version = Info{
		GitVersion: version,
		GitCommit:  commit,
		BuildDate:  date,
	}
}

func SoftwareVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		return bi.Main.Path + "/tree/" + Version.GitVersion
	}
	return Version.GitVersion
}
