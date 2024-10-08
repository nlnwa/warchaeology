package zipfs

import (
	"archive/zip"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/afero"
)

type Fs struct {
	r     *zip.Reader
	files map[string]map[string]*zip.File
}

// splitpath splits a path into directory and file parts.
// It returns the cleaned versions of the directory and file.
// If the path is empty or does not start with a slash, it is prefixed with a slash.
func splitpath(name string) (dir, file string) {
	name = filepath.ToSlash(name)
	if len(name) == 0 || name[0] != os.PathSeparator {
		name = string(os.PathSeparator) + name
	}
	name = filepath.Clean(name)
	dir, file = filepath.Split(name)
	dir = filepath.Clean(dir)
	return
}

func New(r *zip.Reader) afero.Fs {
	fs := &Fs{r: r, files: make(map[string]map[string]*zip.File)}
	for _, file := range r.File {
		d, f := splitpath(file.Name)
		if _, ok := fs.files[d]; !ok {
			fs.files[d] = make(map[string]*zip.File)
		}
		if _, ok := fs.files[d][f]; !ok {
			fs.files[d][f] = file
		}
		if file.FileInfo().IsDir() {
			dirname := filepath.Join(d, f)
			if _, ok := fs.files[dirname]; !ok {
				fs.files[dirname] = make(map[string]*zip.File)
			}
		}

		dv := d
		for {
			d, f = splitpath(dv)
			if _, ok := fs.files[d]; !ok {
				fs.files[d] = make(map[string]*zip.File)
			}
			if f == "" {
				break
			}
			if _, ok := fs.files[d][f]; !ok {
				fs.files[d][f] = nil
			}
			dv = d
		}
	}
	return fs
}

func (fs *Fs) Create(name string) (afero.File, error) { return nil, syscall.EPERM }

func (fs *Fs) Mkdir(name string, perm os.FileMode) error { return syscall.EPERM }

func (fs *Fs) MkdirAll(path string, perm os.FileMode) error { return syscall.EPERM }

func (fs *Fs) Open(name string) (afero.File, error) {
	d, f := splitpath(name)
	if f == "" {
		return &File{fs: fs, isdir: true}, nil
	}
	if _, ok := fs.files[d]; !ok {
		return nil, &os.PathError{Op: "stat", Path: name, Err: syscall.ENOENT}
	}
	file, ok := fs.files[d][f]
	if !ok {
		return nil, &os.PathError{Op: "stat", Path: name, Err: syscall.ENOENT}
	}
	if file == nil {
		return &File{
			fs:        fs,
			pseudodir: &pseudoDir{path: filepath.Join(d, f)},
			isdir:     true,
		}, nil
	}
	return &File{fs: fs, zipfile: file, isdir: file.FileInfo().IsDir()}, nil
}

func (fs *Fs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag != os.O_RDONLY {
		return nil, syscall.EPERM
	}
	return fs.Open(name)
}

func (fs *Fs) Remove(name string) error { return syscall.EPERM }

func (fs *Fs) RemoveAll(path string) error { return syscall.EPERM }

func (fs *Fs) Rename(oldname, newname string) error { return syscall.EPERM }

type pseudoRoot struct{}

func (p *pseudoRoot) Name() string       { return string(filepath.Separator) }
func (p *pseudoRoot) Size() int64        { return 0 }
func (p *pseudoRoot) Mode() os.FileMode  { return os.ModeDir | os.ModePerm }
func (p *pseudoRoot) ModTime() time.Time { return time.Now() }
func (p *pseudoRoot) IsDir() bool        { return true }
func (p *pseudoRoot) Sys() interface{}   { return nil }

type pseudoDir struct {
	path string
}

func (fi pseudoDir) Name() string       { return filepath.Base(fi.path) }
func (fi pseudoDir) Size() int64        { return 0 }
func (fi pseudoDir) IsDir() bool        { return true }
func (fi pseudoDir) ModTime() time.Time { return time.Now() }
func (fi pseudoDir) Mode() os.FileMode  { return os.ModeDir | os.ModePerm }
func (fi pseudoDir) Sys() interface{}   { return nil }

func (fs *Fs) Stat(name string) (os.FileInfo, error) {
	d, f := splitpath(name)
	if f == "" {
		return &pseudoRoot{}, nil
	}
	if _, ok := fs.files[d]; !ok {
		return nil, &os.PathError{Op: "stat", Path: name, Err: syscall.ENOENT}
	}
	file, ok := fs.files[d][f]
	if !ok {
		return nil, &os.PathError{Op: "stat", Path: name, Err: syscall.ENOENT}
	}
	if file == nil {
		return pseudoDir{name}, nil
	}
	return file.FileInfo(), nil
}

func (fs *Fs) Name() string { return "zipfs" }

func (fs *Fs) Chmod(name string, mode os.FileMode) error { return syscall.EPERM }

func (fs *Fs) Chown(name string, uid, gid int) error { return syscall.EPERM }

func (fs *Fs) Chtimes(name string, atime time.Time, mtime time.Time) error { return syscall.EPERM }
