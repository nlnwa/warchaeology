package ls

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal"
	"github.com/nlnwa/warchaeology/internal/utils"
	"github.com/nlnwa/whatwg-url/url"
)

type RecordWriter struct {
	sep        string
	fields     []*field
	sizeFields []*field
}

type toInt64Fn func(wr gowarc.WarcRecord, fileName string, offset int64) int64
type toStringFn func(wr gowarc.WarcRecord, fileName string, offset int64) string
type writerFn func(wr gowarc.WarcRecord, fileName string, offset int64) string

type field struct {
	name   byte
	length int
	align  int
	fn     writerFn
}

func NewRecordWriter(format, separator string) (*RecordWriter, error) {
	recordWriter := &RecordWriter{
		sep: separator,
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

// FormatRecord writes the configured fields for the record to a string. Size is written with a place holder
// since it is not available until the next record is read.
func (recordWriter *RecordWriter) FormatRecord(wr gowarc.WarcRecord, fileName string, offset int64) string {
	stringBuilder := &strings.Builder{}
	for index, field := range recordWriter.fields {
		if index > 0 {
			stringBuilder.WriteString(recordWriter.sep)
		}
		stringBuilder.WriteString(field.fn(wr, fileName, offset))
	}
	return stringBuilder.String()
}

// Write takes a string produced by FormatRecord, replaces eventual size place holders with actual values and
// writes it to stdout.
func (recordWriter *RecordWriter) Write(line string, size int64) {
	var v []interface{}
	for _, sizeField := range recordWriter.sizeFields {
		if sizeField.length > 0 && sizeField.align != 0 {
			v = append(v, sizeField.length)
		}
		v = append(v, size)
	}
	fmt.Printf(line, v...)
	fmt.Println()
}

func createInt64Fn(align, length int, valueFn toInt64Fn) writerFn {
	if length > 0 {
		switch {
		case align < 0:
			return func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%-*d", length, valueFn(warcRecord, fileName, offset))
			}
		case align > 0:
			return func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%*d", length, valueFn(warcRecord, fileName, offset))
			}
		default:
			return func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%d", valueFn(warcRecord, fileName, offset))
			}
		}
	} else {
		return func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return fmt.Sprintf("%d", valueFn(warcRecord, fileName, offset))
		}
	}
}

func createStringFn(align, length int, valueFn toStringFn) writerFn {
	if length > 0 {
		switch {
		case align < 0:
			return func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%-*s", length, utils.CropString(valueFn(warcRecord, fileName, offset), length))
			}
		case align > 0:
			return func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%*s", length, utils.CropString(valueFn(warcRecord, fileName, offset), length))
			}
		default:
			return func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
				return utils.CropString(valueFn(warcRecord, fileName, offset), length)
			}
		}
	} else {
		return func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return valueFn(warcRecord, fileName, offset)
		}
	}
}

func (recordWriter *RecordWriter) createFieldFunc(t *field) {
	switch t.name {
	case 'a':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return warcRecord.WarcHeader().Get(gowarc.WarcTargetURI)
		})
	case 'b':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			if s, err := internal.To14(warcRecord.WarcHeader().Get(gowarc.WarcDate)); err == nil {
				return s
			}
			return "              "
		})
	case 'B':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return warcRecord.WarcHeader().Get(gowarc.WarcDate)
		})
	case 'e':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return warcRecord.WarcHeader().Get(gowarc.WarcIPAddress)
		})
	case 'g':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return fileName
		})
	case 'h':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			if url, err := url.Parse(warcRecord.WarcHeader().Get(gowarc.WarcTargetURI)); err == nil {
				return url.Hostname()
			}
			return ""
		})
	case 'i':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return warcRecord.WarcHeader().Get(gowarc.WarcRecordID)
		})
	case 'k':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return warcRecord.WarcHeader().Get(gowarc.WarcBlockDigest)
		})
	case 'm':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			if httpResponseBlock, ok := warcRecord.Block().(gowarc.HttpResponseBlock); ok {
				return httpResponseBlock.HttpHeader().Get(gowarc.ContentType)
			}
			if httpRequestBlock, ok := warcRecord.Block().(gowarc.HttpRequestBlock); ok {
				return httpRequestBlock.HttpHeader().Get(gowarc.ContentType)
			}
			return ""
		})
	case 'M':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return "-"
		})
	case 'N':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return "-"
		})
	case 'r':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return "-"
		})
	case 's':
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			if httpResponseBlock, ok := warcRecord.Block().(gowarc.HttpResponseBlock); ok {
				return strconv.Itoa(httpResponseBlock.HttpStatusCode())
			}
			return "   "
		})
	case 'S':
		// Size has special handling since value can't be calculated before next record is read.
		t.fn = func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
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
		t.fn = createStringFn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) string {
			return warcRecord.Type().String()
		})
	case 'V':
		t.fn = createInt64Fn(t.align, t.length, func(warcRecord gowarc.WarcRecord, fileName string, offset int64) int64 {
			return offset
		})
	}
}
