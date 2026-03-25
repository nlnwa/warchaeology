package nedlibreader

import (
	"bufio"
	"bytes"
	"io"
	"iter"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/nlnwa/gowarc/v3"
	"github.com/spf13/afero"
)

type NedlibReader struct {
	fs                afero.Fs
	metaFilename      string
	defaultTime       time.Time
	warcRecordOptions []gowarc.WarcRecordOption
	done              bool
}

func NewNedlibReader(fileSystem afero.Fs, metaFilename string, defaultTime time.Time, warcRecordOptions ...gowarc.WarcRecordOption) (*NedlibReader, error) {
	nedlibReader := &NedlibReader{
		fs:                fileSystem,
		metaFilename:      metaFilename,
		defaultTime:       defaultTime,
		warcRecordOptions: warcRecordOptions,
	}
	return nedlibReader, nil
}

func (nedlibReader *NedlibReader) Next() (gowarc.Record, error) {
	var validation []error
	if nedlibReader.done {
		return gowarc.Record{}, io.EOF
	}
	defer func() {
		nedlibReader.done = true
	}()

	file, err := nedlibReader.fs.Open(nedlibReader.metaFilename)
	if err != nil {
		return gowarc.Record{}, err
	}
	defer func() { _ = file.Close() }()

	response, err := http.ReadResponse(bufio.NewReader(io.MultiReader(file, bytes.NewReader([]byte{'\r', '\n'}))), nil)
	if err != nil {
		return gowarc.Record{}, err
	}
	defer response.Body.Close()

	warcRecordBuilder := gowarc.NewRecordBuilder(gowarc.Response, nedlibReader.warcRecordOptions...)

	warcRecordBuilder.AddWarcHeader(gowarc.ContentType, "application/http;msgtype=response")

	header := response.Header

	var warcDate time.Time
	// Try to parse the Date header as a time.Time
	dateString := header.Get("Date")
	if dateString != "" {
		warcDate, _ = parseTime(dateString)
	}
	// Try to parse a path segment as a time.Time
	if warcDate.IsZero() {
		// if one of the path segments is a date, use that as the date (at 12:00 noon)
		segments := strings.Split(nedlibReader.metaFilename, string(filepath.Separator))
		re := regexp.MustCompile(`\d{4}-\d{1,2}-\d{1,2}`)
		for _, dateString := range segments {
			if len(dateString) < 8 {
				continue
			}
			warcDate, err = time.Parse("2006-1-2", re.FindString(dateString))
			if err == nil {
				warcDate = warcDate.Add(time.Hour * 12)
				break
			}
		}
	}
	// Fall back to the default
	if warcDate.IsZero() {
		warcDate = nedlibReader.defaultTime
	}
	warcRecordBuilder.AddWarcHeaderTime(gowarc.WarcDate, warcDate)

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
		return gowarc.Record{}, err
	}
	if err = header.Write(warcRecordBuilder); err != nil {
		return gowarc.Record{}, err
	}

	if _, err = warcRecordBuilder.WriteString("\r\n"); err != nil {
		return gowarc.Record{}, err
	}

	payloadFile, err := nedlibReader.fs.Open(strings.TrimSuffix(nedlibReader.metaFilename, ".meta"))
	if err != nil {
		return gowarc.Record{}, err
	}
	defer func() { _ = payloadFile.Close() }()

	if _, err = warcRecordBuilder.ReadFrom(payloadFile); err != nil {
		return gowarc.Record{}, err
	}

	var warcRecord gowarc.WarcRecord
	warcRecord, validation, err = warcRecordBuilder.Build()
	return gowarc.Record{WarcRecord: warcRecord, Offset: 0, Size: 0, Validation: validation}, err
}

func (nedlibReader *NedlibReader) Records() iter.Seq2[gowarc.Record, error] {
	return func(yield func(gowarc.Record, error) bool) {
		for {
			record, err := nedlibReader.Next()
			if err == io.EOF {
				return
			}
			if !yield(record, err) {
				return
			}
			if err != nil {
				return
			}
		}
	}
}

func (nedlibReader *NedlibReader) Close() error {
	return nil
}
