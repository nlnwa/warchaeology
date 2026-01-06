package arcreader

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/klauspost/compress/gzip"
	mytime "github.com/nationallibraryofnorway/warchaeology/v4/internal/time"
	"github.com/nlnwa/gowarc/v2"
)

type unmarshaler struct {
	opts       []gowarc.WarcRecordOption
	LastOffset int64
	version    int
	gz         *gzip.Reader
}

func NewUnmarshaler(opts ...gowarc.WarcRecordOption) gowarc.Unmarshaler {
	u := &unmarshaler{
		opts: opts,
	}
	return u
}

func (u *unmarshaler) Unmarshal(b *bufio.Reader) (gowarc.WarcRecord, int64, *gowarc.Validation, error) {
	isGzip, r, offset, err := u.searchNextRecord(b)
	if err == io.EOF {
		return nil, offset, nil, err
	}
	if err != nil {
		return nil, offset, nil, fmt.Errorf("could not parse ARC record: %w", err)
	}

	defer func() {
		// Discarding 1 byte which makes up the end of record marker (\n)
		var lf byte = '\n'
		bb, e := r.Peek(4)
		if len(bb) == 0 {
			err = fmt.Errorf("wrong peek: %d, %w", len(bb), e)
		} else {
			if len(bb) != 1 || bb[0] != lf || (e != nil && e != io.EOF) {
				err = fmt.Errorf("wrong peek: %d, %q, %w", len(bb), bb[0], e)
			}
			_, _ = r.Discard(1)
		}

		if isGzip {
			// Empty gzip reader to ensure gzip checksum is validated
			b := make([]byte, 10)
			var err error
			for err == nil {
				_, err = u.gz.Read(b)
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
					return
				}
			}
			_ = u.gz.Close()
		}
	}()

	l, err := r.ReadString('\n')
	if err != nil {
		return nil, 0, nil, fmt.Errorf("could not parse ARC record: %w", err)
	}

	var validation *gowarc.Validation
	var wr gowarc.WarcRecord
	if strings.HasPrefix(l, "filedesc://") {
		wr, validation, err = u.parseFileHeader(r, l)
	} else {
		wr, validation, err = u.parseRecord(r, l)
	}

	return wr, offset, validation, err
}

func (u *unmarshaler) searchNextRecord(b *bufio.Reader) (bool, *bufio.Reader, int64, error) {
	var offset int64
	isGzip := false
	var r *bufio.Reader
	var err error

	// Search for start of new record
	expectedRecordStartOffset := offset
	found := false

	for !found {
		var magic []byte
		magic, err = b.Peek(4)
		if err != nil {
			return false, nil, offset, err
		}

		switch {
		case magic[0] == 0x1f && magic[1] == 0x8b:
			if u.gz == nil {
				u.gz, err = gzip.NewReader(b)
			} else {
				err = u.gz.Reset(b)
			}
			if err != nil {
				if _, err = b.Discard(1); err != nil {
					return false, nil, offset, err
				}
				offset++
				continue
			}
			u.gz.Multistream(false)
			r = bufio.NewReader(u.gz)
			isGzip = true
			found = true

		case bytes.HasPrefix(magic, []byte("http")):
			fallthrough
		case bytes.HasPrefix(magic, []byte("file")):
			fallthrough
		case bytes.HasPrefix(magic, []byte("dns")):
			fallthrough
		case bytes.HasPrefix(magic, []byte("ftp")):
			r = b
			found = true

		default:
			if _, err = b.Discard(1); err != nil {
				return false, nil, offset, err
			}
			offset++
		}
	}

	if expectedRecordStartOffset != offset {
		err = fmt.Errorf("expected start of record at offset: %d, but record was found at offset: %d",
			expectedRecordStartOffset, offset)
	}

	return isGzip, r, offset, err
}

func (u *unmarshaler) parseFileHeader(r *bufio.Reader, l1 string) (gowarc.WarcRecord, *gowarc.Validation, error) {
	var read int
	l2, err := r.ReadString('\n')
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse ARC file header")
	}
	read += len(l2)
	i := strings.IndexByte(l2, ' ')
	v, err := strconv.Atoi(l2[:i])
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse version from ARC file header: %w", err)
	}
	u.version = v

	var recordType gowarc.RecordType
	var url, contentType string
	var date time.Time
	var length int64

	switch u.version {
	case 1:
		recordType, url, _, date, contentType, length, err = u.parseUrlRecordV1(l1)
		if err != nil || recordType != gowarc.Warcinfo {
			return nil, nil, err
		}
	default:
		return nil, nil, fmt.Errorf("unknown ARC record version: %d", v)
	}

	l3, err := r.ReadString('\n')
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse ARC record: %w", err)
	}
	read += len(l3)
	remaining := length - int64(read)

	rb := gowarc.NewRecordBuilder(gowarc.Metadata, u.opts...)
	rb.AddWarcHeader(gowarc.WarcTargetURI, url)
	rb.AddWarcHeader(gowarc.ContentType, contentType)
	rb.AddWarcHeaderTime(gowarc.WarcDate, date)

	if _, err = rb.WriteString(l1 + l2 + l3); err != nil {
		_ = rb.Close()
		return nil, nil, err
	}

	c2 := NewLimitedCountingReader(r, remaining)
	_, err = rb.ReadFrom(c2)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			err = io.EOF
		}
		_ = rb.Close()
		return nil, nil, err
	}

	wr, validation, err := rb.Build()

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
		return nil, nil, fmt.Errorf("unknown ARC record version: %d", u.version)
	}

	rb := gowarc.NewRecordBuilder(0, u.opts...)
	rb.SetRecordType(recordType)
	rb.AddWarcHeader(gowarc.WarcTargetURI, url)
	rb.AddWarcHeader(gowarc.ContentType, contentType)
	rb.AddWarcHeaderTime(gowarc.WarcDate, date)
	rb.AddWarcHeaderInt64(gowarc.ContentLength, length)
	rb.AddWarcHeader(gowarc.WarcIPAddress, ip)

	c2 := NewLimitedCountingReader(r, length)
	_, err = rb.ReadFrom(c2)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			err = io.EOF
		}
		_ = rb.Close()
		return nil, nil, err
	}

	return rb.Build()
}

func (u *unmarshaler) parseUrlRecordV1(l string) (gowarc.RecordType, string, string, time.Time, string, int64, error) {
	reg := regexp.MustCompile(`([^ ]*) ([^ ]*) (\d*) ([^ ]*) (\d*)`)
	subs := reg.FindStringSubmatch(l)
	if len(subs) < 4 {
		return 0, "", "", time.Time{}, "", 0, fmt.Errorf("could not parse ARC record from: %s", l)
	}
	url := subs[1]
	ip := subs[2]
	d := subs[3]
	date, err := mytime.From14ToTime(d)
	if err != nil {
		return 0, "", "", time.Time{}, "", 0, err
	}
	contentType := subs[4]
	length, err := strconv.ParseInt(subs[5], 10, 64)
	if err != nil {
		return 0, "", "", time.Time{}, "", 0, fmt.Errorf("could not parse ARC record: %w", err)
	}

	recordType := gowarc.Response

	switch {
	case strings.HasPrefix(url, "http"):
		contentType = "application/http;msgtype=response"
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
