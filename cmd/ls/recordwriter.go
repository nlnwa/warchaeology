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

package ls

import (
	"fmt"
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal"
	"github.com/nlnwa/whatwg-url/url"
	"regexp"
	"strconv"
	"strings"
)

type RecordWriter struct {
	sep        string
	fields     []*field
	off        int64
	line       string
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

func NewRecordWriter(format, separator string) *RecordWriter {
	rw := &RecordWriter{
		sep: separator,
	}

	re := regexp.MustCompile("([abBeghikmMNrsSTV])([+-]?)(\\d*)")
	m := re.FindAllStringSubmatch(format, -1)
	for _, sm := range m {
		t := &field{name: sm[1][0]}
		if sm[2] == "-" {
			t.align = -1
		}
		if sm[2] == "+" {
			t.align = 1
		}
		if len(sm[3]) > 0 {
			n, err := strconv.ParseInt(sm[3], 10, 32)
			if err != nil {
				panic(err)
			}
			t.length = int(n)
		}
		rw.createFieldFunc(t)
		rw.fields = append(rw.fields, t)
	}

	return rw
}

func (rw *RecordWriter) Write(wr gowarc.WarcRecord, fileName string, offset, size int64) error {
	if rw.line != "" {
		var v []interface{}
		for _, sf := range rw.sizeFields {
			if sf.length > 0 && sf.align != 0 {
				v = append(v, sf.length)
			}
			v = append(v, size)
		}
		fmt.Printf(rw.line, v...)
		fmt.Println()
		rw.line = ""
	}
	if wr != nil {
		s := &strings.Builder{}
		for i, t := range rw.fields {
			if i > 0 {
				s.WriteString(rw.sep)
			}
			s.WriteString(t.fn(wr, fileName, offset))
		}
		rw.line = s.String()
	}
	return nil
}

func createInt64Fn(align, length int, valueFn toInt64Fn) writerFn {
	if length > 0 {
		switch {
		case align < 0:
			l := length
			return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%-*d", l, valueFn(wr, fileName, offset))
			}
		case align > 0:
			l := length
			return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%*d", l, valueFn(wr, fileName, offset))
			}
		default:
			return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%d", valueFn(wr, fileName, offset))
			}
		}
	} else {
		return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return fmt.Sprintf("%d", valueFn(wr, fileName, offset))
		}
	}
}

func createStringFn(align, length int, valueFn toStringFn) writerFn {
	if length > 0 {
		switch {
		case align < 0:
			l := length
			return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%-*s", l, internal.CropString(valueFn(wr, fileName, offset), length))
			}
		case align > 0:
			l := length
			return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%*s", l, internal.CropString(valueFn(wr, fileName, offset), length))
			}
		default:
			return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%s", internal.CropString(valueFn(wr, fileName, offset), length))
			}
		}
	} else {
		return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return fmt.Sprintf("%s", valueFn(wr, fileName, offset))
		}
	}
}

func (rw *RecordWriter) createFieldFunc(t *field) {
	switch t.name {
	case 'a':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return wr.WarcHeader().Get(gowarc.WarcTargetURI)
		})
	case 'b':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			if s, err := internal.To14(wr.WarcHeader().Get(gowarc.WarcDate)); err == nil {
				return s
			}
			return "              "
		})
	case 'B':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return wr.WarcHeader().Get(gowarc.WarcDate)
		})
	case 'e':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return wr.WarcHeader().Get(gowarc.WarcIPAddress)
		})
	case 'g':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return fileName
		})
	case 'h':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			if u, err := url.Parse(wr.WarcHeader().Get(gowarc.WarcTargetURI)); err == nil {
				return u.Hostname()
			}
			return ""
		})
	case 'i':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return wr.WarcHeader().Get(gowarc.WarcRecordID)
		})
	case 'k':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return wr.WarcHeader().Get(gowarc.WarcBlockDigest)
		})
	case 'm':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			if v, ok := wr.Block().(gowarc.HttpResponseBlock); ok {
				return v.HttpHeader().Get(gowarc.ContentType)
			}
			if v, ok := wr.Block().(gowarc.HttpRequestBlock); ok {
				return v.HttpHeader().Get(gowarc.ContentType)
			}
			return ""
		})
	case 'M':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return "-"
		})
	case 'N':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return "-"
		})
	case 'r':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return "-"
		})
	case 's':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			if v, ok := wr.Block().(gowarc.HttpResponseBlock); ok {
				return strconv.Itoa(v.HttpStatusCode())
			}
			return "   "
		})
	case 'S':
		// Size has special handling since value can't be calculated before next record is read.
		t.fn = func(wr gowarc.WarcRecord, fileName string, offset int64) string {
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
		rw.sizeFields = append(rw.sizeFields, t)
	case 'T':
		t.fn = createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return wr.Type().String()
		})
	case 'V':
		t.fn = createInt64Fn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) int64 {
			return offset
		})
	}
}
