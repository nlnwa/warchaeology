package filewalker

import (
	"context"
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

func (fw *FileWalker) Walk(ctx context.Context, path string, walkFn func(fs afero.Fs, path string, err error) error) error {
	return fw.walkDir(ctx, path, path, walkFn)
}

func (fw *FileWalker) walkDir(ctx context.Context, root string, dirName string, walkFn func(fs afero.Fs, path string, err error) error) error {
	return afero.Walk(fw.Fs, dirName, func(path string, info fs.FileInfo, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			return walkFn(fw.Fs, path, err)
		}

		if info.IsDir() {
			// skip already processed directories
			if fw.processedPaths.Contains(path) {
				return filepath.SkipDir
			} else {
				fw.processedPaths.Add(path)
			}
			// always process the path that is equal to the root directory
			if root == path {
				return nil
			}
			// skip directories when recursive option is not set
			if !fw.Recursive {
				return filepath.SkipDir
			}
			return nil
		}

		// handle symlink
		if !info.Mode().IsRegular() {
			if !fw.FollowSymlinks {
				return nil
			}
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
			return fw.walkDir(ctx, root, linkPath, walkFn)
		}
		// filter files by suffix
		if !fw.hasSuffix(path) {
			return nil
		}
		// skip already processed files
		if fw.processedPaths.Contains(path) {
			return nil
		} else {
			fw.processedPaths.Add(path)
		}

		return walkFn(fw.Fs, path, nil)
	})
}
