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

package internal

import (
	"github.com/nlnwa/gowarc"
	"strconv"
)

type Cdx struct {
	// uri (required) - The value should be the non-transformed URI used for the searchable URI (first sortable field).
	Uri string `json:"uri,omitempty"`
	// sha (recommended) - A Base32 encoded SHA-1 digest of the payload that this record refers to. Omit if the URI has
	// no intrinsic payload. For revisit records, this is the digest of the original payload. The algorithm prefix
	// (e.g. sha-1) is not included in this field. See dig for alternative hashing algorithms.
	Sha string `json:"sha,omitempty"`
	// dig - A Base32 encoded output of a hashing algorithm applied to the URI’s payload. This should include a prefix
	// indicating the algorithm.
	Dig string `json:"dig,omitempty"`
	// hsc - HTTP Status Code. Applicable for response records for HTTP(S) URIs.
	Hsc string `json:"hsc,omitempty"`
	// mct - Media Content Type (MIME type). For HTTP(S) response records this is typically the “Content-Type” from
	// the HTTP header. This field, however, does not specify the origin of the information. It may be used to include
	// content type that was derived from content analysis or other sources.
	Mct string `json:"mct,omitempty"`
	// ref (required) - A URI that resolves to the resource that this record refers to. This can be any well defined
	// URI scheme. For the most common web archive use case of warc filename plus offset, see Appendix C. For other
	// use cases, existing schemes can be used or new ones devised.
	Ref string `json:"ref,omitempty"`
	// rid (recommended) - Record ID. Typically WARC-Record-ID or equivalent if not using WARCs. In a mixed environment,
	// you should ensure that record ID is unique.
	Rid string `json:"rid,omitempty"`
	// cle - Content Length. The length of the content (uncompressed), ignoring WARC headers, but including any
	// HTTP headers or similar.
	Cle string `json:"cle,omitempty"`
	// ple - Payload Length. The length of the payload (uncompressed). The exact meaning will vary by content type,
	// but the common case is the length of the document, excluding any HTTP headers in a HTTP response record.
	Ple string `json:"ple,omitempty"`
	// rle - Record Length. The length of the record that this line refers to. This is the entire record
	// (including e.g. WARC headers) as written on disk (compressed if stored compressed).
	Rle string `json:"rle,omitempty"`
	// rct - Record Concurrant To. The record ID of another record that the current record is considered to be
	// ‘concurrant’ to. See further WARC chapter 5.7 (WARC-Concurrent-To).
	Rct string `json:"rct,omitempty"`
	// rou (recommended) - Revisit Original URI. Only valid for records of type revisit. Contains the URI of
	// the record that this record is considered a revisit of.
	Rou string `json:"rou,omitempty"`
	// rod (recommended) - Revisit Original Date. Only valid for records of type revisit. Contains the
	// timestamp (equivalent to sortable field #2) of the record that this record is considered a revisit of.
	Rod string `json:"rod,omitempty"`
	// roi - Revisit Original record ID. Only valid for records of type revisit. Contains the record ID of
	// the record that this record is considered a revisit of.
	Roi string `json:"roi,omitempty"`
	// Searchable URI - ssu (sortable searchable URI)
	Ssu string `json:"ssu,omitempty"`
	// Timestamp - sts (sortable timestamp)
	Sts string `json:"sts,omitempty"`
	// Record Type - srt (sortable record type)
	Srt string `json:"srt,omitempty"`
}

func NewCdxRecord(wr gowarc.WarcRecord, fileName string, offset int64) *Cdx {
	cdx := &Cdx{
		Uri: wr.WarcHeader().Get(gowarc.WarcTargetURI),
		Sha: wr.WarcHeader().Get(gowarc.WarcPayloadDigest),
		Dig: wr.WarcHeader().Get(gowarc.WarcPayloadDigest),
		Ref: "warcfile:" + fileName + "#" + strconv.FormatInt(offset, 10),
		Rid: wr.WarcHeader().Get(gowarc.WarcRecordID),
		Cle: wr.WarcHeader().Get(gowarc.ContentLength),
		//Rle: wr.WarcHeader().Get(warcrecord.ContentLength),
		Rct: wr.WarcHeader().Get(gowarc.WarcConcurrentTo),
		Rou: wr.WarcHeader().Get(gowarc.WarcRefersToTargetURI),
		Rod: wr.WarcHeader().Get(gowarc.WarcRefersToDate),
		Roi: wr.WarcHeader().Get(gowarc.WarcRefersTo),
	}
	if ssu, err := SsurtString(wr.WarcHeader().Get(gowarc.WarcTargetURI), true); err == nil {
		cdx.Ssu = ssu
	}
	cdx.Sts, _ = To14(wr.WarcHeader().Get(gowarc.WarcDate))
	cdx.Srt = wr.Type().String()

	switch v := wr.Block().(type) {
	case gowarc.HttpResponseBlock:
		cdx.Hsc = strconv.Itoa(v.HttpStatusCode())
		cdx.Mct = v.HttpHeader().Get("Content-Type")
		cdx.Ple = v.HttpHeader().Get("Content-Length")
		//case *gowarc.RevisitBlock:
		//	if resp, err := v.Response(); err == nil {
		//		cdx.Hsc = strconv.Itoa(resp.StatusCode)
		//		cdx.Mct = resp.Header.Get("Content-Type")
		//		cdx.Ple = resp.Header.Get("Content-Length")
		//	}
	}

	return cdx
}
