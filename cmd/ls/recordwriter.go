package ls

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v3/internal/time"
	"github.com/nlnwa/warchaeology/v3/internal/util"
	"github.com/nlnwa/warchaeology/v3/internal/warc"
)

type JSONWriter struct {
	logger *log.Logger
	fields []byte
}

func NewJSONWriter(w io.Writer, format string) *JSONWriter {
	var fields []byte
	if format == "" {
		format = "V+11iT-8a100"
	}

	pattern := regexp.MustCompile(`([abBeghikmMNrsSTV])`)
	matches := pattern.FindAllStringSubmatch(format, -1)
	for _, subMatch := range matches {
		name := subMatch[1][0]
		fields = append(fields, name)
	}
	return &JSONWriter{
		fields: fields,
		logger: log.New(w, "", 0),
	}
}

func (recordWriter *JSONWriter) WriteRecord(record warc.Record, fileName string) error {
	metadata := warc.Metadata{}
	warcRecord := record.WarcRecord
	for _, field := range recordWriter.fields {
		switch field {
		case 'a':
			metadata.Url = warc.Url(warcRecord)
		case 'b', 'B':
			metadata.Date, _ = warc.Date(warcRecord)
		case 'e':
			metadata.IpAddress = warc.IpAddress(warcRecord)
		case 'g':
			metadata.FileName = fileName
		case 'h':
			metadata.Hostname = warc.Hostname(warcRecord)
		case 'i':
			metadata.RecordId = warc.RecordId(warcRecord)
		case 'k':
			metadata.Checksum = warc.Checksum(warcRecord)
		case 'm':
			metadata.MimeType = warc.MimeType(warcRecord)
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

	recordWriter.logger.Println(string(b))
	return nil
}

type RecordWriter struct {
	w          io.Writer
	fields     []*field
	sizeFields []*field
	sep        string
	logger     *log.Logger
}

type toInt64Fn func(record warc.Record, file string) int64
type toStringFn func(record warc.Record, file string) string
type writerFn func(record warc.Record, file string) string

type field struct {
	name   byte
	length int
	align  int
	fn     writerFn
}

func NewRecordWriter(w io.Writer, format, separator string) (*RecordWriter, error) {
	recordWriter := &RecordWriter{
		w:      w,
		sep:    separator,
		logger: log.New(w, "", 0),
	}
	if format == "" {
		format = "V+11iT-8a100"
	}
	pattern := regexp.MustCompile(`([abBeghikmMNrsSTV])([+-]?)(\d*)`)
	matches := pattern.FindAllStringSubmatch(format, -1)
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

func (recordWriter *RecordWriter) WriteRecord(record warc.Record, fileName string) error {
	line := recordWriter.FormatRecord(record, fileName)
	return recordWriter.Write(line, record.Size)
}

// FormatRecord writes the configured fields for the record to a string. Size is written with a place holder
// since it is not available until the next record is read.
func (recordWriter *RecordWriter) FormatRecord(record warc.Record, fileName string) string {
	stringBuilder := &strings.Builder{}
	for index, field := range recordWriter.fields {
		if index > 0 {
			stringBuilder.WriteString(recordWriter.sep)
		}
		stringBuilder.WriteString(field.fn(record, fileName))
	}
	return stringBuilder.String()
}

// Write takes a string produced by FormatRecord, replaces eventual size place holders with actual values and
// writes it to stdout.
func (recordWriter *RecordWriter) Write(line string, size int64) error {
	var v []interface{}
	for _, sizeField := range recordWriter.sizeFields {
		if sizeField.length > 0 && sizeField.align != 0 {
			v = append(v, sizeField.length)
		}
		v = append(v, size)
	}
	recordWriter.logger.Printf(line+"\n", v...)
	return nil
}

func createInt64Fn(align, length int, valueFn toInt64Fn) writerFn {
	if length > 0 {
		switch {
		case align < 0:
			return func(record warc.Record, file string) string {
				return fmt.Sprintf("%-*d", length, valueFn(record, file))
			}
		case align > 0:
			return func(record warc.Record, file string) string {
				return fmt.Sprintf("%*d", length, valueFn(record, file))
			}
		default:
			return func(record warc.Record, file string) string {
				return fmt.Sprintf("%d", valueFn(record, file))
			}
		}
	} else {
		return func(record warc.Record, file string) string {
			return fmt.Sprintf("%d", valueFn(record, file))
		}
	}
}

func createStringFn(align, length int, valueFn toStringFn) writerFn {
	if length > 0 {
		switch {
		case align < 0:
			return func(record warc.Record, file string) string {
				return fmt.Sprintf("%-*s", length, util.CropString(valueFn(record, file), length))
			}
		case align > 0:
			return func(record warc.Record, file string) string {
				return fmt.Sprintf("%*s", length, util.CropString(valueFn(record, file), length))
			}
		default:
			return func(record warc.Record, file string) string {
				return util.CropString(valueFn(record, file), length)
			}
		}
	} else {
		return func(record warc.Record, file string) string {
			return valueFn(record, file)
		}
	}
}

// createFieldFunc creates a function for a field based on the field name.
func (recordWriter *RecordWriter) createFieldFunc(t *field) {
	// The following fields are supported:
	// a - original URL
	// b - date in 14 digit format
	// B - date in RFC3339 format
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
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return warc.Url(record.WarcRecord)
		})
	case 'b':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			if s, err := time.To14(record.WarcRecord.WarcHeader().Get(gowarc.WarcDate)); err == nil {
				return s
			}
			return "              "
		})
	case 'B':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			t, err := warc.Date(record.WarcRecord)
			if err == nil {
				return time.UTCW3cIso8601(t)
			}
			return "                         "
		})
	case 'e':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return warc.IpAddress(record.WarcRecord)
		})
	case 'g':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return file
		})
	case 'h':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return warc.Hostname(record.WarcRecord)
		})
	case 'i':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return warc.RecordId(record.WarcRecord)
		})
	case 'k':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return warc.Checksum(record.WarcRecord)
		})
	case 'm':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return warc.MimeType(record.WarcRecord)
		})
	case 'M':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return "-"
		})
	case 'N':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return "-"
		})
	case 'r':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return "-"
		})
	case 's':
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			statusCode := warc.StatusCode(record.WarcRecord)
			if statusCode > 0 {
				return strconv.Itoa(statusCode)
			}
			return "   "
		})
	case 'S':
		// Size has special handling since value can't be calculated before next record is read.
		t.fn = func(record warc.Record, file string) string {
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
		t.fn = createStringFn(t.align, t.length, func(record warc.Record, file string) string {
			return record.WarcRecord.Type().String()
		})
	case 'V':
		t.fn = createInt64Fn(t.align, t.length, func(record warc.Record, file string) int64 {
			return record.Offset
		})
	}
}
