package filewalker

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type FileWalker struct {
	Recursive      bool
	FollowSymlinks bool
	Suffixes       []string
	Fs             afero.Fs
	processedPaths StringSet
}

func WithRecursive(recursive bool) func(*FileWalker) {
	return func(w *FileWalker) {
		w.Recursive = recursive
	}
}

func WithFollowSymlinks(followSymlinks bool) func(*FileWalker) {
	return func(w *FileWalker) {
		w.FollowSymlinks = followSymlinks
	}
}

func WithSuffixes(suffixes []string) func(*FileWalker) {
	return func(w *FileWalker) {
		w.Suffixes = suffixes
	}
}

func WithFs(fs afero.Fs) func(*FileWalker) {
	return func(w *FileWalker) {
		w.Fs = fs
	}
}

func New(options ...func(*FileWalker)) *FileWalker {
	opts := &FileWalker{
		processedPaths: NewStringSet(),
	}
	for _, opt := range options {
		opt(opts)
	}
	return opts
}

func (fw *FileWalker) hasSuffix(path string) bool {
	if len(fw.Suffixes) == 0 {
		return true
	}
	for _, suffix := range fw.Suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

func (fw *FileWalker) Walk(path string, walkFn func(fs afero.Fs, path string, err error) error) error {
	return fw.walkDir(path, path, walkFn)
}

func (fw *FileWalker) walkDir(root string, dirName string, walkFn func(fs afero.Fs, path string, err error) error) error {
	return afero.Walk(fw.Fs, dirName, func(path string, info fs.FileInfo, err error) error {
		// handle error
		if err != nil {
			return walkFn(fw.Fs, path, err)
		}

		// handle directory
		if info.IsDir() {
			if fw.processedPaths.Contains(path) {
				// skip directory if already processed
				return filepath.SkipDir
			}
			fw.processedPaths.Add(path)
			if root == path {
				// always process an initial directory
				return nil
			}
			if !fw.Recursive {
				// skip directory if not recursive option is set
				return filepath.SkipDir
			}
			return nil
		}

		// handle symlink
		if fw.FollowSymlinks && !info.Mode().IsRegular() {
			linkReader, ok := fw.Fs.(afero.LinkReader)
			if !ok {
				return afero.ErrNoReadlink
			}
			linkPath, err := linkReader.ReadlinkIfPossible(path)
			if err != nil {
				return err
			}
			if !filepath.IsAbs(linkPath) {
				linkPath = filepath.Join(filepath.Dir(path), linkPath)
			}
			return fw.walkDir(root, linkPath, walkFn)
		}
		if !info.Mode().IsRegular() {
			// skip non-regular files
			return nil
		}
		if !fw.hasSuffix(path) {
			// skip file if suffix does not match
			return nil
		}
		if fw.processedPaths.Contains(path) {
			// skip file if already processed
			return nil
		}
		fw.processedPaths.Add(path)
		return walkFn(fw.Fs, path, nil)
	})
}
