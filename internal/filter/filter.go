/*
 * Copyright 2022 National Library of Norway.
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
 *
 */

package filter

import (
	"math"
	"strconv"
	"strings"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/utils"
)

type Filter struct {
	ids         []string
	RecordTypes gowarc.RecordType
	fromStatus  int
	toStatus    int
	mime        []string
}

func New(options ...func(*Filter)) *Filter {
	f := &Filter{}
	for _, option := range options {
		option(f)
	}
	return f
}

func WithRecordIds(ids []string) func(*Filter) {
	return func(f *Filter) {
		for id := range ids {
			f.ids = append(f.ids, strings.Trim(ids[id], "<>"))
		}
	}
}

func WithRecordTypes(recordTypes []string) func(*Filter) {
	return func(f *Filter) {
		for _, r := range recordTypes {
			switch strings.ToLower(r) {
			case "warcinfo":
				f.RecordTypes = f.RecordTypes | gowarc.Warcinfo
			case "request":
				f.RecordTypes = f.RecordTypes | gowarc.Request
			case "response":
				f.RecordTypes = f.RecordTypes | gowarc.Response
			case "metadata":
				f.RecordTypes = f.RecordTypes | gowarc.Metadata
			case "revisit":
				f.RecordTypes = f.RecordTypes | gowarc.Revisit
			case "resource":
				f.RecordTypes = f.RecordTypes | gowarc.Resource
			case "continuation":
				f.RecordTypes = f.RecordTypes | gowarc.Continuation
			case "conversion":
				f.RecordTypes = f.RecordTypes | gowarc.Conversion
			}
		}
	}
}

func WithResponseCode(responseCode string) func(*Filter) {
	return func(f *Filter) {
		rc := strings.Split(responseCode, "-")
		switch len(rc) {
		case 1:
			if len(rc[0]) == 0 {
				f.toStatus = math.MaxInt32
			} else {
				if i, e := strconv.Atoi(rc[0]); e == nil {
					f.fromStatus = i
					f.toStatus = i + 1
				} else {
					panic(e)
				}
			}
		case 2:
			if len(rc[0]) > 0 {
				if i, e := strconv.Atoi(rc[0]); e == nil {
					f.fromStatus = i
				} else {
					panic(e)
				}
			}
			if len(rc[1]) == 0 {
				f.toStatus = math.MaxInt32
			} else {
				if i, e := strconv.Atoi(rc[1]); e == nil {
					f.toStatus = i
				} else {
					panic(e)
				}
			}
		default:
			panic("Illegal response code")
		}
	}
}

func WithMimeType(mimeTypes []string) func(*Filter) {
	return func(f *Filter) {
		f.mime = mimeTypes
	}
}

func (f *Filter) Accept(wr gowarc.WarcRecord) bool {
	// Check record ID's
	if len(f.ids) > 0 {
		if !utils.Contains(f.ids, wr.RecordId()) {
			return false
		}
	}

	// Check Record type
	if f.RecordTypes != 0 && wr.Type()&f.RecordTypes == 0 {
		return false
	}

	// Check HTTP response code
	if v, ok := wr.Block().(gowarc.HttpResponseBlock); ok {
		if v.HttpStatusCode() < f.fromStatus || v.HttpStatusCode() >= f.toStatus {
			return false
		}
	}

	// Check document mime-type
	if len(f.mime) > 0 {
		switch v := wr.Block().(type) {
		case gowarc.HttpRequestBlock:
			for _, r := range f.mime {
				if strings.Contains(v.HttpHeader().Get(gowarc.ContentType), r) {
					return true
				}
			}
			return false
		case gowarc.HttpResponseBlock:
			for _, r := range f.mime {
				if strings.Contains(v.HttpHeader().Get(gowarc.ContentType), r) {
					return true
				}
			}
			return false
		default:
			return false
		}
	}

	return true
}
