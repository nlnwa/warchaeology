/*
 * Copyright 2021 National Library of Norway.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nlnwa/warchaeology/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	// Find 'this' directory relative to this file to allow callers to be in any package
	var _, b, _, _ = runtime.Caller(0)
	var dir = filepath.Dir(b)
	var docsDir = filepath.Join(dir, "../../docs/content")

	genMdDoc(docsDir)
}

const fmTemplate = `---
date: %s
title: "%s"
slug: %s
url: %s
---
`

func genMdDoc(docsDir string) {
	var dir = filepath.Join(docsDir, "cmd")
	fmt.Println("generating documentation")

	if files, err := os.ReadDir(dir); err == nil {
		for _, f := range files {
			if strings.HasPrefix(f.Name(), "warc") {
				p := path.Join(dir, f.Name())
				_ = os.Remove(p)
			}
		}
	}

	c := cmd.NewCommand()
	c.DisableAutoGenTag = true

	filePrepender := func(filename string) string {
		now := time.Now().Format(time.RFC3339)
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, path.Ext(name))
		url := "/cmd/" + strings.ToLower(base) + "/"
		return fmt.Sprintf(fmTemplate, now, strings.Replace(base, "_", " ", -1), base, url)
	}

	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		return "../" + strings.ToLower(base) + "/"
	}

	if err := doc.GenMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
		panic(err)
	}
}
