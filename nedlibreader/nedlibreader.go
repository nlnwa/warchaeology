package nedlibreader

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nlnwa/gowarc"
	"github.com/spf13/afero"
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
	defer response.Body.Close()

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
