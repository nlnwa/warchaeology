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
 */

package arcreader

import (
	"io"
	"sync/atomic"
)

// CountingReader counts the bytes read through it.
type CountingReader struct {
	bytesRead int64
	maxBytes  int64
	ioReader  io.Reader
}

// NewCountingReader makes a new CountingReader that counts the bytes
// read through it.
func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{
		ioReader: r,
		maxBytes: -1,
	}
}

// NewLimitedCountingReader makes a new CountingReader that counts the bytes
// read through it.
//
// When maxBytes bytes are read, the next read will
// return io.EOF even though the underlying reader has more data.
func NewLimitedCountingReader(r io.Reader, maxBytes int64) *CountingReader {
	return &CountingReader{
		ioReader: r,
		maxBytes: maxBytes,
	}
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	if r.maxBytes >= 0 {
		remaining := r.maxBytes - r.N()
		if remaining <= 0 {
			return 0, io.EOF
		}
		if int64(len(p)) > remaining {
			p = p[:remaining]
		}
		n, err = r.ioReader.Read(p)
		atomic.AddInt64(&r.bytesRead, int64(n))
	} else {
		n, err = r.ioReader.Read(p)
		atomic.AddInt64(&r.bytesRead, int64(n))
	}
	return
}

// N gets the number of bytes that have been read
// so far.
func (r *CountingReader) N() int64 {
	return atomic.LoadInt64(&r.bytesRead)
}
