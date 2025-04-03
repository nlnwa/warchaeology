package version

import (
	"fmt"
	"runtime"
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
`, v.GitVersion, v.GitCommit, v.BuildDate, v.GoVersion, v.Compiler, v.Platform)
}

func Set(version, commit, date string) {
	Version = Info{
		GitVersion: version,
		GitCommit:  commit,
		BuildDate:  date,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func SoftwareVersion() string {
	return fmt.Sprintf("%s %s", "Warchaeology", Version.GitVersion)
}
