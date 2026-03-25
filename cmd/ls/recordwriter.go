package ls

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/nationallibraryofnorway/warchaeology/v5/internal/time"
	"github.com/nationallibraryofnorway/warchaeology/v5/internal/util"
	"github.com/nationallibraryofnorway/warchaeology/v5/internal/warc"
	"github.com/nlnwa/gowarc/v3"
)

var fieldPattern = regexp.MustCompile(`([abBeghikmMNrsSTV])([+-]?)(\d*)`)

func selectedFields(format string) []byte {
	if format == "" {
		format = "V+11iT-8a100"
	}

	matches := fieldPattern.FindAllStringSubmatch(format, -1)
	fields := make([]byte, 0, len(matches))
	for _, subMatch := range matches {
		fields = append(fields, subMatch[1][0])
	}
	return fields
}

func fieldsNeedParsedBlock(format string) bool {
	for _, field := range selectedFields(format) {
		switch field {
		case 'm', 's':
			return true
		}
	}
	return false
}

type JSONWriter struct {
	w      io.Writer
	fields []byte
}

func NewJSONWriter(w io.Writer, format string) *JSONWriter {
	fields := selectedFields(format)
	return &JSONWriter{
		fields: fields,
		w:      w,
	}
}

func (recordWriter *JSONWriter) WriteRecord(record gowarc.Record, fileName string) error {
	metadata := warc.Metadata{}
	warcRecord := record.WarcRecord
	for _, field := range recordWriter.fields {
		switch field {
		case 'a':
			metadata.Url = warc.URL(warcRecord)
		case 'b':
			date, err := warc.Date(warcRecord)
			if err != nil {
				return fmt.Errorf("failed to parse date: %w", err)
			}
			metadata.Date = time.To14(date)
		case 'B':
			date, err := warc.Date(warcRecord)
			if err != nil {
				return fmt.Errorf("failed to parse date: %w", err)
			}
			metadata.Date = time.ToW3CDTF(date)
		case 'e':
			metadata.IpAddress = warc.IPAddress(warcRecord)
		case 'g':
			metadata.FileName = fileName
		case 'h':
			metadata.Hostname = warc.Hostname(warcRecord)
		case 'i':
			metadata.RecordId = warc.RecordID(warcRecord)
		case 'k':
			metadata.Checksum = warc.Checksum(warcRecord)
		case 'm':
			metadata.MimeType = warc.MIMEType(warcRecord)
		case 'M':
		case 'N':
		case 'r':
		case 's':
			metadata.StatusCode = warc.StatusCode(warcRecord)
		case 'S':
			metadata.Size = record.Size
		case 'T':
			metadata.Type = warcRecord.Type().String()
		case 'V':
			metadata.Offset = record.Offset
		}
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = recordWriter.w.Write(b)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(recordWriter.w)
	return err
}

type RecordWriter struct {
	w          io.Writer
	fields     []*field
	sizeFields []*field
	sep        string
}

type toInt64Fn func(record gowarc.Record, file string) int64
type toStringFn func(record gowarc.Record, file string) string
type writerFn func(record gowarc.Record, file string) string

type field struct {
	name   byte
	length int
	align  int
	fn     writerFn
}

func NewRecordWriter(w io.Writer, format string, separator string) (*RecordWriter, error) {
	recordWriter := &RecordWriter{
		w:   w,
		sep: separator,
	}
	if format == "" {
		format = "V+11iT-8a100"
	}
	matches := fieldPattern.FindAllStringSubmatch(format, -1)
	for _, subMatch := range matches {
		field := &field{name: subMatch[1][0]}
		if subMatch[2] == "-" {
			field.align = -1
		}
		if subMatch[2] == "+" {
			field.align = 1
		}
		if len(subMatch[3]) > 0 {
			length, err := strconv.ParseInt(subMatch[3], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("failed to parse field length, original error: '%w'", err)
			}
			field.length = int(length)
		}
		recordWriter.createFieldFunc(field)
		recordWriter.fields = append(recordWriter.fields, field)
	}

	return recordWriter, nil
}

func (recordWriter *RecordWriter) WriteRecord(record gowarc.Record, fileName string) error {
	line := recordWriter.FormatRecord(record, fileName)
	return recordWriter.Write(line, record.Size)
}

// FormatRecord writes the configured fields for the record to a string.
func (recordWriter *RecordWriter) FormatRecord(record gowarc.Record, fileName string) string {
	var stringBuilder strings.Builder
	stringBuilder.Grow(len(recordWriter.fields) * 16)
	for index, field := range recordWriter.fields {
		if index > 0 {
			_, _ = stringBuilder.WriteString(recordWriter.sep)
		}
		_, _ = stringBuilder.WriteString(field.fn(record, fileName))
	}
	return stringBuilder.String()
}

// Write takes a string produced by FormatRecord, replaces eventual size place holders with actual values and
// writes it to stdout.
func (recordWriter *RecordWriter) Write(line string, size int64) error {
	if len(recordWriter.sizeFields) == 0 {
		if _, err := io.WriteString(recordWriter.w, line); err != nil {
			return err
		}
		_, err := io.WriteString(recordWriter.w, "\n")
		return err
	}

	var v []any
	for _, sizeField := range recordWriter.sizeFields {
		if sizeField.length > 0 && sizeField.align != 0 {
			v = append(v, sizeField.length)
		}
		v = append(v, size)
	}
	_, err := fmt.Fprintf(recordWriter.w, line+"\n", v...)
	return err
}

func padString(value string, align, length int) string {
	if length <= 0 || len(value) >= length {
		return value
	}
	padding := strings.Repeat(" ", length-len(value))
	if align < 0 {
		return value + padding
	}
	if align > 0 {
		return padding + value
	}
	return value
}

func formatInt(value int64, align, length int) string {
	return padString(strconv.FormatInt(value, 10), align, length)
}

func formatText(value string, align, length int) string {
	if length > 0 {
		value = util.CropString(value, length)
	}
	return padString(value, align, length)
}

func createInt64Fn(align, length int, valueFn toInt64Fn) writerFn {
	return func(record gowarc.Record, file string) string {
		return formatInt(valueFn(record, file), align, length)
	}
}

func createStringFn(align, length int, valueFn toStringFn) writerFn {
	return func(record gowarc.Record, file string) string {
		return formatText(valueFn(record, file), align, length)
	}
}

// createFieldFunc creates a function for a field based on the field name.
func (recordWriter *RecordWriter) createFieldFunc(t *field) {
	// The following fields are supported:
	// a - original URL
	// b - date in 14 digit format
	// B - date in RFC3339 format (up to 9 fractional digits)
	// e - IP
	// g - file name
	// h - original host
	// i - record id
	// k - checksum
	// m - document mime type
	// s - http response code
	// S - record size in WARC file
	// T - record type
	// V - Offset in WARC file
	switch t.name {
	case 'a':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return warc.URL(record.WarcRecord)
		})
	case 'b':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			t, err := warc.Date(record.WarcRecord)
			if err != nil {
				return "              "
			}
			return time.To14(t)
		})
	case 'B':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			t, err := warc.Date(record.WarcRecord)
			if err != nil {
				return "                   "
			}
			return time.ToW3CDTF(t)
		})
	case 'e':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return warc.IPAddress(record.WarcRecord)
		})
	case 'g':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return file
		})
	case 'h':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return warc.Hostname(record.WarcRecord)
		})
	case 'i':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return warc.RecordID(record.WarcRecord)
		})
	case 'k':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return warc.Checksum(record.WarcRecord)
		})
	case 'm':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return warc.MIMEType(record.WarcRecord)
		})
	case 'M':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return "-"
		})
	case 'N':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return "-"
		})
	case 'r':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return "-"
		})
	case 's':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			statusCode := warc.StatusCode(record.WarcRecord)
			if statusCode > 0 {
				return strconv.Itoa(statusCode)
			}
			return "   "
		})
	case 'S':
		// Size has special handling since value can't be calculated before next record is read.
		t.fn = func(record gowarc.Record, file string) string {
			if t.length > 0 {
				switch {
				case t.align < 0:
					return "%-*d"
				case t.align > 0:
					return "%*d"
				default:
					return "%d"
				}
			} else {
				return "%d"
			}
		}
		recordWriter.sizeFields = append(recordWriter.sizeFields, t)
	case 'T':
		t.fn = createStringFn(t.align, t.length, func(record gowarc.Record, file string) string {
			return record.WarcRecord.Type().String()
		})
	case 'V':
		t.fn = createInt64Fn(t.align, t.length, func(record gowarc.Record, file string) int64 {
			return record.Offset
		})
	}
}
