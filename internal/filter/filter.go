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
	"github.com/spf13/viper"
	"strings"
)

type Filter struct {
	RecordTypes gowarc.RecordType
}

func NewFromViper() *Filter {
	f := &Filter{}

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

	return f
}

func (f *Filter) Accept(wr gowarc.WarcRecord) bool {
	// Check Record type
	if f.RecordTypes != 0 && wr.Type()&f.RecordTypes == 0 {
		return false
	}

	return true
}
