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
	sep       string
	fields    []writerFn
	off       int64
	line      string
	sizeField int
}

func NewRecordWriter(format, separator string) *RecordWriter {
	c := &RecordWriter{
		sep: separator,
	}
	tokens := parseFormat(format)
	for _, t := range tokens {
		switch t.name {
		case 'a':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return wr.WarcHeader().Get(gowarc.WarcTargetURI)
			})
			c.fields = append(c.fields, f)
		case 'b':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				if s, err := internal.To14(wr.WarcHeader().Get(gowarc.WarcDate)); err == nil {
					return s
				}
				return "              "
			})
			c.fields = append(c.fields, f)
		case 'B':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return wr.WarcHeader().Get(gowarc.WarcDate)
			})
			c.fields = append(c.fields, f)
		case 'e':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return wr.WarcHeader().Get(gowarc.WarcIPAddress)
			})
			c.fields = append(c.fields, f)
		case 'g':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return fileName
			})
			c.fields = append(c.fields, f)
		case 'h':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				if u, err := url.Parse(wr.WarcHeader().Get(gowarc.WarcTargetURI)); err == nil {
					return u.Hostname()
				}
				return ""
			})
			c.fields = append(c.fields, f)
		case 'i':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return wr.WarcHeader().Get(gowarc.WarcRecordID)
			})
			c.fields = append(c.fields, f)
		case 'k':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return wr.WarcHeader().Get(gowarc.WarcBlockDigest)
			})
			c.fields = append(c.fields, f)
		case 'm':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				if v, ok := wr.Block().(gowarc.HttpResponseBlock); ok {
					return v.HttpHeader().Get(gowarc.ContentType)
				}
				if v, ok := wr.Block().(gowarc.HttpRequestBlock); ok {
					return v.HttpHeader().Get(gowarc.ContentType)
				}
				return ""
			})
			c.fields = append(c.fields, f)
		case 'M':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return "-"
			})
			c.fields = append(c.fields, f)
		case 'N':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return "-"
			})
			c.fields = append(c.fields, f)
		case 'r':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return "-"
			})
			c.fields = append(c.fields, f)
		case 's':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				if v, ok := wr.Block().(gowarc.HttpResponseBlock); ok {
					return strconv.Itoa(v.HttpStatusCode())
				}
				return "   "
			})
			c.fields = append(c.fields, f)
		case 'S':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return "%d"
			})
			c.fields = append(c.fields, f)
			c.sizeField++
		case 'T':
			f := createStringFn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return wr.Type().String()
			})
			c.fields = append(c.fields, f)
		case 'V':
			f := createInt64Fn(t.align, t.length, func(wr gowarc.WarcRecord, fileName string, offset int64) int64 {
				return offset
			})
			c.fields = append(c.fields, f)
		}
	}

	return c
}

type toInt64Fn func(wr gowarc.WarcRecord, fileName string, offset int64) int64
type toStringFn func(wr gowarc.WarcRecord, fileName string, offset int64) string

type writerFn func(wr gowarc.WarcRecord, fileName string, offset int64) string

func (c *RecordWriter) Write(wr gowarc.WarcRecord, fileName string, offset int64) error {
	size := offset - c.off
	c.off = offset
	if c.line != "" {
		var sf []interface{}
		for i := 0; i < c.sizeField; i++ {
			sf = append(sf, size)
		}
		fmt.Printf(c.line, sf...)
		fmt.Println()
		c.line = ""
	}
	if wr != nil {
		s := &strings.Builder{}
		for i, fn := range c.fields {
			if i > 0 {
				s.WriteString(c.sep)
			}
			s.WriteString(fn(wr, fileName, offset))
		}
		c.line = s.String()
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
		case align == 0:
			return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%d", valueFn(wr, fileName, offset))
			}
		}
	} else {
		return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return fmt.Sprintf("%d", valueFn(wr, fileName, offset))
		}
	}
	return nil
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
		case align == 0:
			return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
				return fmt.Sprintf("%s", internal.CropString(valueFn(wr, fileName, offset), length))
			}
		}
	} else {
		return func(wr gowarc.WarcRecord, fileName string, offset int64) string {
			return fmt.Sprintf("%s", valueFn(wr, fileName, offset))
		}
	}
	return nil
}

type token struct {
	name   byte
	length int
	align  int
}

func parseFormat(format string) []*token {
	var res []*token

	re := regexp.MustCompile("([abBeghikmMNrsSTV])([+-]?)(\\d*)")
	m := re.FindAllStringSubmatch(format, -1)
	for _, sm := range m {
		t := &token{name: sm[1][0]}
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
		res = append(res, t)
	}

	return res
}
