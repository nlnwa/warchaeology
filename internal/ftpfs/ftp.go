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
	"context"
	"github.com/jackc/puddle"
	"github.com/jlaffaye/ftp"
	"github.com/spf13/afero"
	"os"
	"syscall"
	"time"
)

// Fs is an implementation of afero.Fs that utilizes functions from the ftp package.
//
// For detailed information on any method, please refer to the documentation of the ftp
// package available at github.com/jlaffaye/ftp.
type Fs struct {
	pool *puddle.Pool
}

func New(addr, user, passwd string, poolSize int32) afero.Fs {
	constructor := func(ctx context.Context) (interface{}, error) {
		c, err := ftp.Dial(addr, ftp.DialWithContext(ctx))
		if err != nil {
			return nil, err
		}
		err = c.Login(user, passwd)
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	destructor := func(value interface{}) {
		_ = value.(*ftp.ServerConn).Quit()
	}

	pool := puddle.NewPool(constructor, destructor, poolSize)
	return &Fs{
		pool: pool,
	}
}

func (s Fs) Name() string { return "ftpfs" }

func (s Fs) Create(name string) (afero.File, error) {
	return nil, syscall.EROFS
}

func (s Fs) Mkdir(name string, perm os.FileMode) error {
	return syscall.EROFS
}

func (s Fs) MkdirAll(path string, perm os.FileMode) error {
	return syscall.EROFS
}

func (s Fs) Open(name string) (afero.File, error) {
	res, err := s.pool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	return FileOpen(res, name)
}

func (s Fs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag == os.O_RDONLY {
		return s.Open(name)
	}
	return nil, syscall.EROFS
}

func (s Fs) Remove(name string) error {
	return syscall.EROFS
}

func (s Fs) RemoveAll(path string) error {
	return syscall.EROFS
}

func (s Fs) Rename(oldname, newname string) error {
	return syscall.EROFS
}

func (s Fs) Stat(name string) (os.FileInfo, error) {
	f, err := s.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return f.Stat()
}

func (s Fs) Chmod(name string, mode os.FileMode) error {
	return syscall.EROFS
}

func (s Fs) Chown(name string, uid, gid int) error {
	return syscall.EROFS
}

func (s Fs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return syscall.EROFS
}
