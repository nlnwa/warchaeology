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
