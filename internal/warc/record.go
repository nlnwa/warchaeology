package warc

import (
	"net/url"
	"time"

	"github.com/nlnwa/gowarc/v2"
)

type Metadata struct {
	Url        string    `json:"url,omitempty"`
	Date       time.Time `json:"date,omitempty"`
	IpAddress  string    `json:"ipAddress,omitempty"`
	FileName   string    `json:"filename,omitempty"`
	Hostname   string    `json:"hostname,omitempty"`
	RecordId   string    `json:"recordId,omitempty"`
	Checksum   string    `json:"checksum,omitempty"`
	MimeType   string    `json:"mimeType,omitempty"`
	StatusCode int       `json:"statusCode,omitempty"`
	Size       int64     `json:"size,omitempty"`
	Type       string    `json:"type,omitempty"`
	Offset     int64     `json:"offset,omitempty"`
}

func Url(wr gowarc.WarcRecord) string {
	return wr.WarcHeader().Get(gowarc.WarcTargetURI)
}

func Date(wr gowarc.WarcRecord) (time.Time, error) {
	return wr.WarcHeader().GetTime(gowarc.WarcDate)
}

func IpAddress(wr gowarc.WarcRecord) string {
	return wr.WarcHeader().Get(gowarc.WarcIPAddress)
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

func RecordId(wr gowarc.WarcRecord) string {
	return wr.WarcHeader().GetId(gowarc.WarcRecordID)
}

func Checksum(wr gowarc.WarcRecord) string {
	return wr.WarcHeader().Get(gowarc.WarcBlockDigest)
}

func MimeType(wr gowarc.WarcRecord) string {
	switch block := wr.Block().(type) {
	case gowarc.HttpResponseBlock:
		if block.HttpHeader() == nil {
			return ""
		}
		return block.HttpHeader().Get(gowarc.ContentType)
	case gowarc.HttpRequestBlock:
		if block.HttpHeader() == nil {
			return ""
		}
		return block.HttpHeader().Get(gowarc.ContentType)
	default:
		return ""
	}
}

func StatusCode(wr gowarc.WarcRecord) int {
	if httpResponseBlock, ok := wr.Block().(gowarc.HttpResponseBlock); ok {
		return httpResponseBlock.HttpStatusCode()
	}
	return 0
}
