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
		want    *ref
		wantErr bool
	}{
		{"1",
			[]byte{2, 99, 101, 49, 53, 49, 101, 97, 101, 45, 50, 98, 98, 48, 45, 52, 49, 97, 55, 45, 97, 49, 98, 53,
				45, 97, 57, 56, 52, 100, 53, 101, 52, 102, 97, 55, 48, 1, 0, 0, 0, 14, 188, 239, 152, 159, 0, 0, 0, 0, 255,
				255, 104, 116, 116, 112, 58, 47, 47, 119, 119, 119, 46, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109},
			&ref{gowarc.RevisitRef{
				Profile:        gowarc.ProfileIdenticalPayloadDigestV1_1,
				TargetRecordId: "<urn:uuid:ce151eae-2bb0-41a7-a1b5-a984d5e4fa70>",
				TargetUri:      "http://www.example.com",
				TargetDate:     "2006-11-17T11:48:47Z",
			}},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &ref{}
			err := got.UnmarshalBinary(tt.value)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ref{*tt.value}
			got, err := r.MarshalBinary()
			assert.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encode() got = %s, want %v", got, tt.want)
			}
		})
	}
}

//func Test_index_isRevisit(t *testing.T) {
//	type fields struct {
//		digests *badger.DB
//	}
//	type args struct {
//		key string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    *gowarc.RevisitRef
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			idx := &index{
//				digests: tt.fields.digests,
//			}
//			got, err := idx.isRevisit(tt.args.key)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("isRevisit() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("isRevisit() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
