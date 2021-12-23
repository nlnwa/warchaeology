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
	"github.com/nlnwa/warchaeology/cmd"
	"github.com/spf13/cobra/doc"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	// Find 'this' directory relative to this file to allow callers to be in any package
	var _, b, _, _ = runtime.Caller(0)
	var dir = filepath.Dir(b)
	var docsDir = filepath.Join(dir, "../../docs")
	var templateDir = filepath.Join(dir, "docs")

	if len(os.Args) != 2 {
		panic("Missing version")
	}

	genMdDoc(docsDir)
	parseTemplates(templateDir, docsDir, os.Args[1])
}

func genMdDoc(docsDir string) {
	var dir = filepath.Join(docsDir, "cmd")
	fmt.Println("generating documentation")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(dir, 0777); err != nil {
		panic(err)
	}

	c := cmd.NewCommand()
	c.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(c, dir); err != nil {
		panic(err)
	}
}

func parseTemplates(templateDir, docsDir, version string) {
	err := filepath.WalkDir(templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			relDir, err := filepath.Rel(templateDir, path)
			if err != nil {
				return err
			}
			outDir := filepath.Join(docsDir, relDir)
			return os.MkdirAll(outDir, 0777)
		} else if strings.HasSuffix(path, "tmpl") {
			if filepath.Ext(d.Name()) != ".tmpl" {
				return nil
			}
			outFile := strings.TrimSuffix(d.Name(), ".tmpl")
			relDir, err := filepath.Rel(templateDir, filepath.Dir(path))
			if err != nil {
				return err
			}
			outPath := filepath.Join(docsDir, relDir, outFile)
			fmt.Printf("docsDir: %s, path: %s, file: %s, rel: %s, out: %s\n", docsDir, path, d.Name(), relDir, outPath)
			t, err := template.ParseFiles(path)
			if err != nil {
				return err
			}
			out, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer out.Close()

			return t.Execute(out, version)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
