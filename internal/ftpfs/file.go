/*
 * Copyright 2024 National Library of Norway.
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

package ftpfs

import (
	"github.com/jackc/puddle"
	"github.com/jlaffaye/ftp"
	"io"
	"os"
	"syscall"
)

type File struct {
	resp         *ftp.Response
	info         *FileInfo
	resource     *puddle.Resource
	actualPos    int64
	requestedPos int64
}

func FileOpen(res *puddle.Resource, name string) (*File, error) {
	f := &File{resource: res}
	entry, err := f.ftpClient().List(name)
	if err != nil {
		return nil, err
	}
	if len(entry) == 1 {
		f.info = &FileInfo{entry: entry[0], fullName: name}
	} else {
		f.info = &FileInfo{entry: &ftp.Entry{
			Name: name,
			Type: ftp.EntryTypeFolder,
			Size: 0,
		},
			fullName: name,
		}
	}

	return f, nil
}

func (f *File) Close() error {
	if f.resp != nil {
		_ = f.resp.Close()
	}
	if f.resource != nil {
		f.resource.Release()
		f.resource = nil
	}
	return nil
}

func (f *File) ftpClient() *ftp.ServerConn {
	return f.resource.Value().(*ftp.ServerConn)
}

func (f *File) Name() string {
	return f.info.Name()
}

func (f *File) Stat() (os.FileInfo, error) {
	return f.info, nil
}

func (f *File) Sync() error {
	return nil
}

func (f *File) Truncate(size int64) error {
	return syscall.EROFS
}

func (f *File) Read(b []byte) (n int, err error) {
	if f.actualPos != f.requestedPos {
		return f.ReadAt(b, f.requestedPos)
	}

	if f.resp == nil {
		resp, err := f.ftpClient().Retr(f.info.fullName)
		if err != nil {
			return 0, err
		}
		f.resp = resp
	}
	n, err = f.resp.Read(b)
	f.actualPos += int64(n)
	f.requestedPos = f.actualPos
	return
}

func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	if f.resp != nil {
		f.resp.Close()
		f.resp = nil
	}

	resp, err := f.ftpClient().RetrFrom(f.info.fullName, uint64(off))
	if err != nil {
		return 0, err
	}
	f.resp = resp

	n, err = f.resp.Read(b)
	f.actualPos = off + int64(n)
	f.requestedPos = f.actualPos
	return
}

func (f *File) Readdir(count int) (res []os.FileInfo, err error) {
	w := f.ftpClient().Walk(f.info.fullName)
	for w.Next() {
		res = append(res, &FileInfo{entry: w.Stat()})
	}
	if err := w.Err(); err != nil {
		return nil, err
	}
	return
}

func (f *File) Readdirnames(n int) (names []string, err error) {
	data, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}
	for _, v := range data {
		names = append(names, v.Name())
	}
	return
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.requestedPos = offset
	case io.SeekCurrent:
		f.requestedPos = f.actualPos + offset
	case io.SeekEnd:
		f.requestedPos = f.info.Size() - offset
	default:
		return 0, syscall.ENOTSUP
	}
	return f.requestedPos, nil
}

func (f *File) Write(b []byte) (n int, err error) {
	return 0, syscall.EROFS
}

func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	return 0, syscall.EROFS
}

func (f *File) WriteString(s string) (ret int, err error) {
	return 0, syscall.EROFS
}
