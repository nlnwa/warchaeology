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
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal"
	log "github.com/sirupsen/logrus"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Unmarshaler interface {
	Unmarshal(b *bufio.Reader) (gowarc.WarcRecord, int64, *gowarc.Validation, error)
}

type unmarshaler struct {
	opts       []gowarc.WarcRecordOption
	LastOffset int64
	version    int
	warcInfoID string
}

func NewUnmarshaler(opts ...gowarc.WarcRecordOption) Unmarshaler {
	u := &unmarshaler{
		opts: opts,
	}
	return u
}

func (u *unmarshaler) Unmarshal(b *bufio.Reader) (gowarc.WarcRecord, int64, *gowarc.Validation, error) {
	var r *bufio.Reader
	var offset int64
	validation := &gowarc.Validation{}

	magic, err := b.Peek(5)
	if err != nil {
		return nil, offset, validation, err
	}

	var g *gzip.Reader
	if magic[0] == 0x1f && magic[1] == 0x8b {
		log.Debug("detected gzip record")
		var x io.ByteReader
		x = io.ByteReader(b)
		g, err = gzip.NewReader(x.(io.Reader))
		if err != nil {
			return nil, offset, validation, err
		}
		g.Multistream(false)
		r = bufio.NewReader(g)
	} else {
		fmt.Printf("Magic: %x %x\n", magic[0], magic[1])
		r = b
	}

	l, err := r.ReadString('\n')
	if err != nil {
		return nil, 0, nil, fmt.Errorf("Could not parse ARC record: %w", err)
	}

	var wr gowarc.WarcRecord
	if strings.HasPrefix(l, "filedesc://") {
		wr, validation, err = u.parseFileHeader(r, l)
	} else {
		wr, validation, err = u.parseRecord(r, l)
		if err != nil {
			fmt.Printf("STREAM %T %v\n", r, g)
		}
	}

	// Discarding 1 byte which makes up the end of record marker (\n)
	// TODO: validate that record ends with correct marker
	_, _ = r.Discard(1)
	if g != nil {
		n, err := io.Copy(io.Discard, g)
		if n > 0 {
			fmt.Println("AFTER READ", n, err)
			panic("JADDA")
		}
		_ = g.Close()
	}
	return wr, 0, validation, err
}

func (u *unmarshaler) parseFileHeader(r *bufio.Reader, l1 string) (gowarc.WarcRecord, *gowarc.Validation, error) {
	var read int
	l2, err := r.ReadString('\n')
	if err != nil {
		return nil, nil, fmt.Errorf("Could not parse ARC file header")
	}
	read += len(l2)
	i := strings.IndexByte(l2, ' ')
	v, err := strconv.Atoi(l2[:i])
	if err != nil {
		return nil, nil, fmt.Errorf("Could not parse version from ARC file header: %w", err)
	}
	u.version = v

	var recordType gowarc.RecordType
	var url, contentType string
	var date time.Time
	var length int64

	switch u.version {
	case 1:
		recordType, url, _, date, contentType, length, err = u.parseUrlRecordV1(l1)
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, fmt.Errorf("Uknown ARC record version: %d", v)
	}

	l3, err := r.ReadString('\n')
	if err != nil {
		return nil, nil, fmt.Errorf("Could not parse ARC record: %w", err)
	}
	read += len(l3)
	//fmt.Println(l3)
	remaining := length - int64(read)

	rb := gowarc.NewRecordBuilder(0, u.opts...)
	rb.SetRecordType(recordType)
	rb.AddWarcHeader(gowarc.WarcFilename, strings.TrimPrefix(url, "filedesc://"))
	rb.AddWarcHeader(gowarc.ContentType, contentType)

	rb.AddWarcHeaderTime(gowarc.WarcDate, date)
	rb.AddWarcHeaderInt64(gowarc.ContentLength, remaining)

	c2 := NewLimitedCountingReader(r, remaining)
	rb.ReadFrom(c2)

	wr, validation, err := rb.Build()
	if wr.Type() == gowarc.Warcinfo {
		u.warcInfoID = wr.WarcHeader().Get(gowarc.WarcRecordID)
	}
	return wr, validation, err
}

func (u *unmarshaler) parseRecord(r *bufio.Reader, l1 string) (gowarc.WarcRecord, *gowarc.Validation, error) {
	var recordType gowarc.RecordType
	var url, ip, contentType string
	var date time.Time
	var length int64
	var err error

	switch u.version {
	case 1:
		recordType, url, ip, date, contentType, length, err = u.parseUrlRecordV1(l1)
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, fmt.Errorf("Uknown ARC record version: %d", u.version)
	}

	rb := gowarc.NewRecordBuilder(0, u.opts...)
	rb.SetRecordType(recordType)
	rb.AddWarcHeader(gowarc.WarcTargetURI, url)
	rb.AddWarcHeader(gowarc.ContentType, contentType)
	rb.AddWarcHeaderTime(gowarc.WarcDate, date)
	rb.AddWarcHeaderInt64(gowarc.ContentLength, length)
	rb.AddWarcHeader(gowarc.WarcIPAddress, ip)
	rb.AddWarcHeader(gowarc.WarcWarcinfoID, u.warcInfoID)

	c2 := NewLimitedCountingReader(r, length)
	_, err = rb.ReadFrom(c2)
	if err != nil {
		return nil, nil, err
	}

	return rb.Build()
}

func (u *unmarshaler) parseUrlRecordV1(l string) (gowarc.RecordType, string, string, time.Time, string, int64, error) {
	reg := regexp.MustCompile("([^ ]*) ([^ ]*) (\\d*) ([^ ]*) (\\d*)")
	subs := reg.FindStringSubmatch(l)
	if subs == nil || len(subs) < 4 {
		return 0, "", "", time.Time{}, "", 0, fmt.Errorf("Could not parse ARC record from: %s", l)
	}
	url := subs[1]
	ip := subs[2]
	d := subs[3]
	date, err := internal.From14ToTime(d)
	if err != nil {
		return 0, "", "", time.Time{}, "", 0, err
	}
	contentType := subs[4]
	length, err := strconv.ParseInt(subs[5], 10, 64)
	if err != nil {
		return 0, "", "", time.Time{}, "", 0, fmt.Errorf("Could not parse ARC record: %w", err)
	}

	recordType := gowarc.Response

	switch {
	case strings.HasPrefix(url, "http"):
		contentType = "application/http; msgtype=response"
	case strings.HasPrefix(url, "dns:"):
		contentType = "text/dns"
		recordType = gowarc.Resource
	case strings.HasPrefix(url, "filedesc://"):
		recordType = gowarc.Warcinfo
	default:
		recordType = gowarc.Resource
	}

	return recordType, url, ip, date, contentType, length, nil
}
