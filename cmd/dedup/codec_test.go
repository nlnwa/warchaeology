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

package dedup

import (
	"github.com/nlnwa/gowarc"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_codec_decode(t *testing.T) {
	tests := []struct {
		name    string
		value   []byte
		want    *gowarc.RevisitRef
		wantErr bool
	}{
		{"1",
			[]byte{2, 99, 101, 49, 53, 49, 101, 97, 101, 45, 50, 98, 98, 48, 45, 52, 49, 97, 55, 45, 97, 49, 98, 53,
				45, 97, 57, 56, 52, 100, 53, 101, 52, 102, 97, 55, 48, 1, 0, 0, 0, 14, 188, 239, 152, 159, 0, 0, 0, 0, 255,
				255, 104, 116, 116, 112, 58, 47, 47, 119, 119, 119, 46, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109},
			&gowarc.RevisitRef{
				Profile:        gowarc.ProfileIdenticalPayloadDigestV1_1,
				TargetRecordId: "<urn:uuid:ce151eae-2bb0-41a7-a1b5-a984d5e4fa70>",
				TargetUri:      "http://www.example.com",
				TargetDate:     "2006-11-17T11:48:47Z",
			},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalRevisitRef(tt.value)
			if err != nil {
				panic(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encode()\n got: %s\nwant: %v", got, tt.want)
			}
		})
	}
}

func Test_codec_encode(t *testing.T) {
	tests := []struct {
		name    string
		value   *gowarc.RevisitRef
		want    []byte
		wantErr bool
	}{
		{"1", &gowarc.RevisitRef{
			Profile:        gowarc.ProfileIdenticalPayloadDigestV1_1,
			TargetRecordId: "<urn:uuid:ce151eae-2bb0-41a7-a1b5-a984d5e4fa70>",
			TargetUri:      "http://www.example.com",
			TargetDate:     "2006-11-17T11:48:47Z",
		},
			[]byte{2, 99, 101, 49, 53, 49, 101, 97, 101, 45, 50, 98, 98, 48, 45, 52, 49, 97, 55, 45, 97, 49, 98, 53,
				45, 97, 57, 56, 52, 100, 53, 101, 52, 102, 97, 55, 48, 1, 0, 0, 0, 14, 188, 239, 152, 159, 0, 0, 0, 0, 255,
				255, 104, 116, 116, 112, 58, 47, 47, 119, 119, 119, 46, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109},
			false},
		{"2", &gowarc.RevisitRef{
			Profile:        gowarc.ProfileIdenticalPayloadDigestV1_1,
			TargetRecordId: "urn:uuid:0c54500b-6d1a-43ff-9446-a18a35af9bf6",
			TargetUri:      "https://www.ringebu.kommune.no/css/css.ashx?MId1=12434&MId2=-1&MId3=-1&style=login",
			TargetDate:     "2019-11-13T22:54:15Z",
		},
			[]byte{2, 48, 99, 53, 52, 53, 48, 48, 98, 45, 54, 100, 49, 97, 45, 52, 51, 102, 102, 45, 57, 52, 52, 54,
				45, 97, 49, 56, 97, 51, 53, 97, 102, 57, 98, 102, 54, 1, 0, 0, 0, 14, 213, 94, 128, 151, 0, 0, 0, 0, 255,
				255, 104, 116, 116, 112, 115, 58, 47, 47, 119, 119, 119, 46, 114, 105, 110, 103, 101, 98, 117, 46, 107, 111,
				109, 109, 117, 110, 101, 46, 110, 111, 47, 99, 115, 115, 47, 99, 115, 115, 46, 97, 115, 104, 120, 63, 77, 73,
				100, 49, 61, 49, 50, 52, 51, 52, 38, 77, 73, 100, 50, 61, 45, 49, 38, 77, 73, 100, 51, 61, 45, 49, 38, 115,
				116, 121, 108, 101, 61, 108, 111, 103, 105, 110},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalRevisitRef(tt.value)
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encode() got = %v,\n want %v", got, tt.want)
			}
		})
	}
}
