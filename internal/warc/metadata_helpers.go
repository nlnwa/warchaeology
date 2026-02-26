package warc

import (
	"time"

	"github.com/nlnwa/gowarc/v3"
	"github.com/nlnwa/whatwg-url/url"
)

type Metadata struct {
	Url        string `json:"url,omitempty"`
	Date       string `json:"date,omitempty"`
	IpAddress  string `json:"ipAddress,omitempty"`
	FileName   string `json:"filename,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	RecordId   string `json:"recordId,omitempty"`
	Checksum   string `json:"checksum,omitempty"`
	MimeType   string `json:"mimeType,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
	Size       int64  `json:"size,omitempty"`
	Type       string `json:"type,omitempty"`
	Offset     int64  `json:"offset,omitempty"`
}

func URL(wr gowarc.WarcRecord) string {
	return wr.WarcHeader().Get(gowarc.WarcTargetURI)
}

func Url(wr gowarc.WarcRecord) string {
	return URL(wr)
}

func Date(wr gowarc.WarcRecord) (time.Time, error) {
	return wr.WarcHeader().GetTime(gowarc.WarcDate)
}

func IPAddress(wr gowarc.WarcRecord) string {
	return wr.WarcHeader().Get(gowarc.WarcIPAddress)
}

func IpAddress(wr gowarc.WarcRecord) string {
	return IPAddress(wr)
}

func FileName(wr gowarc.WarcRecord) string {
	return wr.WarcHeader().Get(gowarc.WarcFilename)
}

func Hostname(wr gowarc.WarcRecord) string {
	uri := wr.WarcHeader().Get(gowarc.WarcTargetURI)

	if url, err := url.Parse(uri); err == nil {
		return url.Hostname()
	}
	return ""
}

func RecordID(wr gowarc.WarcRecord) string {
	return wr.WarcHeader().GetId(gowarc.WarcRecordID)
}

func RecordId(wr gowarc.WarcRecord) string {
	return RecordID(wr)
}

func Checksum(wr gowarc.WarcRecord) string {
	return wr.WarcHeader().Get(gowarc.WarcBlockDigest)
}

func MIMEType(wr gowarc.WarcRecord) string {
	switch block := wr.Block().(type) {
	case gowarc.HttpResponseBlock:
		if block.HttpHeader() != nil {
			return block.HttpHeader().Get("Content-Type")
		}
	case gowarc.HttpRequestBlock:
		if block.HttpHeader() == nil {
			return block.HttpHeader().Get("Content-Type")
		}
	}
	return ""
}

func MimeType(wr gowarc.WarcRecord) string {
	return MIMEType(wr)
}

func StatusCode(wr gowarc.WarcRecord) int {
	if httpResponseBlock, ok := wr.Block().(gowarc.HttpResponseBlock); ok {
		return httpResponseBlock.HttpStatusCode()
	}
	return 0
}
