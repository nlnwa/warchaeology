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
	"slices"
	"strings"

	"github.com/nlnwa/gowarc/v2"
)

type RecordFilter struct {
	ids         []string
	RecordTypes gowarc.RecordType
	fromStatus  int
	toStatus    int
	mime        []string
}

func New(options ...func(*RecordFilter)) *RecordFilter {
	f := &RecordFilter{}
	for _, option := range options {
		option(f)
	}
	return f
}

func WithCodeRange(from, to int) func(*RecordFilter) {
	return func(f *RecordFilter) {
		f.fromStatus = from
		f.toStatus = to
	}
}

func WithRecordTypes(recordTypes gowarc.RecordType) func(*RecordFilter) {
	return func(f *RecordFilter) {
		f.RecordTypes = recordTypes
	}
}

func WithRecordIds(ids []string) func(*RecordFilter) {
	return func(f *RecordFilter) {
		for id := range ids {
			f.ids = append(f.ids, strings.Trim(ids[id], "<>"))
		}
	}
}
func WithMimeType(mimeTypes []string) func(*RecordFilter) {
	return func(f *RecordFilter) {
		f.mime = mimeTypes
	}
}

func (f *RecordFilter) Accept(wr gowarc.WarcRecord) bool {
	// Check record ID's
	if len(f.ids) > 0 && !slices.Contains(f.ids, wr.RecordId()) {
		return false
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
			if v.HttpHeader() == nil {
				return false
			}
			contentType := v.HttpHeader().Get(gowarc.ContentType)
			for _, r := range f.mime {
				if strings.Contains(contentType, r) {
					return true
				}
			}
			return false
		case gowarc.HttpResponseBlock:
			if v.HttpHeader() == nil {
				return false
			}
			contentType := v.HttpHeader().Get(gowarc.ContentType)
			for _, r := range f.mime {
				if strings.Contains(contentType, r) {
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
