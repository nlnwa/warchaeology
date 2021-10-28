/*
Copyright © 2019 National Library of Norway

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Command line utility to work with WARC files.
package main

import (
	"fmt"
	"github.com/nlnwa/warchaeology/cmd"
	"runtime"
)

// Overridden by ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	runtime.GOMAXPROCS(128)
	c := cmd.NewCommand()
	c.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s\n" .Version}}`)
	c.Version = fmt.Sprintf("%s (commit: %s, %s)", version, commit, date)
	_ = c.Execute()
}
