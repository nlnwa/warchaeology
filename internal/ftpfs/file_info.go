package ftpfs

import (
	"io/fs"
	"os"
	"time"

	"github.com/jlaffaye/ftp"
)

const defaultFileMode = 0o555

type FileInfo struct {
	entry    *ftp.Entry
	fullName string
}

func (f FileInfo) Name() string {
	return f.entry.Name
}

func (f FileInfo) Size() int64 {
	return int64(f.entry.Size)
}

func (f FileInfo) Mode() fs.FileMode {
	var m os.FileMode = defaultFileMode
	if f.entry.Type == ftp.EntryTypeFolder {
		m = m | os.ModeDir
	}
	if f.entry.Type == ftp.EntryTypeLink {
		m = m | os.ModeSymlink
	}
	return m
}

func (f FileInfo) ModTime() time.Time {
	return f.entry.Time
}

func (f FileInfo) IsDir() bool {
	return f.Mode().IsDir()
}

func (f FileInfo) Sys() any {
	return nil
}
