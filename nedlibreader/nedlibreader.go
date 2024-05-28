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
	fileSystem        afero.Fs
	metaFilename      string
	defaultTime       time.Time
	warcRecordOptions []gowarc.WarcRecordOption
	done              bool
}

func NewNedlibReader(fileSystem afero.Fs, metaFilename string, defaultTime time.Time, warcRecordOptions ...gowarc.WarcRecordOption) (*NedlibReader, error) {
	nedlibReader := &NedlibReader{
		fileSystem:        fileSystem,
		metaFilename:      metaFilename,
		defaultTime:       defaultTime,
		warcRecordOptions: warcRecordOptions,
	}
	return nedlibReader, nil
}

func (nedlibReader *NedlibReader) Next() (gowarc.WarcRecord, int64, *gowarc.Validation, error) {
	var validation *gowarc.Validation
	if nedlibReader.done {
		return nil, 0, validation, io.EOF
	}
	defer func() { nedlibReader.done = true }()

	file, err := nedlibReader.fileSystem.Open(nedlibReader.metaFilename)
	if err != nil {
		return nil, 0, validation, err
	}

	response, err := http.ReadResponse(bufio.NewReader(io.MultiReader(file, bytes.NewReader([]byte{'\r', '\n'}))), nil)
	_ = file.Close()
	if err != nil {
		return nil, 0, validation, err
	}
	defer response.Body.Close()

	warcRecordBuilder := gowarc.NewRecordBuilder(gowarc.Response, nedlibReader.warcRecordOptions...)

	warcRecordBuilder.AddWarcHeader(gowarc.ContentType, "application/http;msgtype=response")

	header := response.Header
	dateString := header.Get("Date")
	if dateString != "" {
		t, err := parseTime(dateString)
		if err != nil {
			return nil, 0, validation, err
		}
		warcRecordBuilder.AddWarcHeaderTime(gowarc.WarcDate, t)
	} else {
		warcRecordBuilder.AddWarcHeaderTime(gowarc.WarcDate, nedlibReader.defaultTime)
	}

	for field := range header {
		if strings.HasPrefix(field, "Arc") {
			switch field {
			case "Arc-Url":
				warcRecordBuilder.AddWarcHeader(gowarc.WarcTargetURI, header.Get(field))
			case "Arc-Length":
				header.Set(gowarc.ContentLength, header.Get(field))
			}
			header.Del(field)
		}
	}
	if _, err = warcRecordBuilder.WriteString(response.Proto + " " + response.Status + "\n"); err != nil {
		return nil, 0, validation, err
	}
	if err = header.Write(warcRecordBuilder); err != nil {
		return nil, 0, validation, err
	}

	if _, err = warcRecordBuilder.WriteString("\r\n"); err != nil {
		return nil, 0, validation, err
	}

	metaFile, err := nedlibReader.fileSystem.Open(strings.TrimSuffix(nedlibReader.metaFilename, ".meta"))
	if err != nil {
		return nil, 0, validation, err
	}
	defer func() { _ = metaFile.Close() }()

	if _, err = warcRecordBuilder.ReadFrom(metaFile); err != nil {
		return nil, 0, validation, err
	}

	var warcRecord gowarc.WarcRecord
	warcRecord, validation, err = warcRecordBuilder.Build()
	return warcRecord, 0, validation, err
}

func (nedlibReader *NedlibReader) Close() error {
	return nil
}
