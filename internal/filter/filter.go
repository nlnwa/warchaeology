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
	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/flag"
	"github.com/nlnwa/warchaeology/internal/utils"
	"github.com/spf13/viper"
	"math"
	"strconv"
	"strings"
)

type Filter struct {
	ids         []string
	RecordTypes gowarc.RecordType
	fromStatus  int
	toStatus    int
	mime        []string
}

func NewFromViper() *Filter {
	f := &Filter{}

	// Parse record ID's flag
	f.ids = viper.GetStringSlice(flag.RecordId)
	for i := 0; i < len(f.ids); i++ {
		f.ids[i] = strings.Trim(f.ids[i], "<>")
	}

	// Parse record types flag
	recordTypes := viper.GetStringSlice(flag.RecordType)
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

	// Parse response code flag
	rc := viper.GetString(flag.ResponseCode)
	responseCodes := strings.Split(rc, "-")
	switch len(responseCodes) {
	case 1:
		if len(responseCodes[0]) == 0 {
			f.toStatus = math.MaxInt32
		} else {
			if i, e := strconv.Atoi(responseCodes[0]); e == nil {
				f.fromStatus = i
				f.toStatus = i + 1
			} else {
				panic(e)
			}
		}
	case 2:
		if len(responseCodes[0]) > 0 {
			if i, e := strconv.Atoi(responseCodes[0]); e == nil {
				f.fromStatus = i
			} else {
				panic(e)
			}
		}
		if len(responseCodes[1]) == 0 {
			f.toStatus = math.MaxInt32
		} else {
			if i, e := strconv.Atoi(responseCodes[1]); e == nil {
				f.toStatus = i
			} else {
				panic(e)
			}
		}
	default:
		panic("Illegal response code")
	}

	// Parse document mmime-type flag
	f.mime = viper.GetStringSlice(flag.MimeType)

	return f
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
