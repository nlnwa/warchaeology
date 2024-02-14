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

package nedlibreader

import (
	"bufio"
	"bytes"
	"github.com/nlnwa/gowarc"
	"github.com/spf13/afero"
	"io"
	"net/http"
	"strings"
	"time"
)

type NedlibReader struct {
	fs           afero.Fs
	metaFilename string
	defaultTime  time.Time
	opts         []gowarc.WarcRecordOption
	done         bool
}

func NewNedlibReader(fs afero.Fs, metaFilename string, defaultTime time.Time, opts ...gowarc.WarcRecordOption) (*NedlibReader, error) {
	n := &NedlibReader{
		fs:           fs,
		metaFilename: metaFilename,
		defaultTime:  defaultTime,
		opts:         opts,
	}
	return n, nil
}

func (n *NedlibReader) Next() (gowarc.WarcRecord, int64, *gowarc.Validation, error) {
	var validation *gowarc.Validation
	if n.done {
		return nil, 0, validation, io.EOF
	}
	defer func() { n.done = true }()

	f, err := n.fs.Open(n.metaFilename)
	if err != nil {
		return nil, 0, validation, err
	}

	response, err := http.ReadResponse(bufio.NewReader(io.MultiReader(f, bytes.NewReader([]byte{'\r', '\n'}))), nil)
	_ = f.Close()
	if err != nil {
		return nil, 0, validation, err
	}

	rb := gowarc.NewRecordBuilder(gowarc.Response, n.opts...)

	rb.AddWarcHeader(gowarc.ContentType, "application/http;msgtype=response")

	header := response.Header
	dateString := header.Get("Date")
	if dateString != "" {
		t, err := time.Parse(time.RFC1123, dateString)
		if err != nil {
			return nil, 0, validation, err
		}
		rb.AddWarcHeaderTime(gowarc.WarcDate, t)
	} else {
		rb.AddWarcHeaderTime(gowarc.WarcDate, n.defaultTime)
	}

	for i := range header {
		if strings.HasPrefix(i, "Arc") {
			switch i {
			case "Arc-Url":
				rb.AddWarcHeader(gowarc.WarcTargetURI, header.Get(i))
			case "Arc-Length":
				header.Set(gowarc.ContentLength, header.Get(i))
			}
			header.Del(i)
		}
	}
	if _, err = rb.WriteString(response.Proto + " " + response.Status + "\n"); err != nil {
		return nil, 0, validation, err
	}
	if err = header.Write(rb); err != nil {
		return nil, 0, validation, err
	}

	if _, err = rb.WriteString("\r\n"); err != nil {
		return nil, 0, validation, err
	}

	p, err := n.fs.Open(strings.TrimSuffix(n.metaFilename, ".meta"))
	if err != nil {
		return nil, 0, validation, err
	}
	defer func() { _ = p.Close() }()

	if _, err = rb.ReadFrom(p); err != nil {
		return nil, 0, validation, err
	}

	var wr gowarc.WarcRecord
	wr, validation, err = rb.Build()
	return wr, 0, validation, err
}

func (n *NedlibReader) Close() error {
	return nil
}
