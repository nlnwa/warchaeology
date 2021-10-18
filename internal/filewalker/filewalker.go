package filewalker

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type FileWalker interface {
	Walk() error
}

type filewalker struct {
	paths          []string
	recursive      bool
	followSymlinks bool
	suffixes       []string
	fn             func(path string)
	processedPaths map[string]bool
}

func New(paths []string, recursive, followSymlinks bool, suffixes []string, fn func(path string)) FileWalker {
	return &filewalker{paths: paths, recursive: recursive, followSymlinks: followSymlinks, suffixes: suffixes, fn: fn, processedPaths: map[string]bool{}}
}

func (f *filewalker) Walk() error {
	for _, p := range f.paths {
		if !f.processedPaths[p] {
			if err := f.walkDir(p, p); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *filewalker) walkDir(root, dirName string) error {
	return filepath.WalkDir(dirName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println("Error:", err)
			return filepath.SkipDir
		}
		if d.IsDir() {
			f.processedPaths[path] = true
			if !f.recursive && root != path {
				return filepath.SkipDir
			}
		} else if !d.IsDir() && !d.Type().IsRegular() {
			if f.followSymlinks {
				s, _ := filepath.EvalSymlinks(path)
				if f.processedPaths[s] {
					return nil
				}
				return f.walkDir(root, s)
			}
		} else if f.hasSuffix(path) {
			f.fn(path)
		}
		return nil
	})
}

func (f *filewalker) hasSuffix(path string) bool {
	if f.suffixes == nil || len(f.suffixes) == 0 {
		return true
	}
	for _, s := range f.suffixes {
		if strings.HasSuffix(path, s) {
			return true
		}
	}
	return false
}
