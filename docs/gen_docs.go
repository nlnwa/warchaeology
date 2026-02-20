package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nationallibraryofnorway/warchaeology/v4/cmd"
	"github.com/spf13/cobra/doc"
)

//go:generate go run ./gen_docs.go
func main() {
	// Find 'this' directory relative to this file to allow callers to be in any package
	var _, b, _, _ = runtime.Caller(0)
	var dir = filepath.Dir(b)
	var docsDir = filepath.Join(dir, "reference", "cli")

	genMdDoc(docsDir)
}

const fmTemplate = `---
title: "%s"
---
`

func genMdDoc(dir string) {
	fmt.Println("Generating documentation...")

	if files, err := os.ReadDir(dir); err == nil {
		for _, f := range files {
			if strings.HasPrefix(f.Name(), "warc") {
				p := filepath.Join(dir, f.Name())
				_ = os.Remove(p)
			}
		}
	}

	c := cmd.NewWarcCommand()
	c.DisableAutoGenTag = true

	filePrepender := func(filename string) string {
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, filepath.Ext(name))
		return fmt.Sprintf(fmTemplate, strings.ReplaceAll(base, "_", " "))
	}

	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		return strings.ToLower(base) + ".md"
	}

	if err := doc.GenMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
		panic(err)
	}
}
