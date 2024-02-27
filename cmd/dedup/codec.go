package dedup

import (
	"strings"
	"time"

	"github.com/nlnwa/gowarc"
)

const (
	oProfile = 0
	oId      = oProfile + 1
	oDate    = oId + 36
	oUri     = oDate + 15
)

func UnmarshalRevisitRef(data []byte) (*gowarc.RevisitRef, error) {
	r := &gowarc.RevisitRef{}
	switch data[0] {
	case 1:
		r.Profile = gowarc.ProfileIdenticalPayloadDigestV1_0
	case 2:
		r.Profile = gowarc.ProfileIdenticalPayloadDigestV1_1
	case 3:
		r.Profile = gowarc.ProfileServerNotModifiedV1_0
	case 4:
		r.Profile = gowarc.ProfileServerNotModifiedV1_1
	}
	r.TargetRecordId = "<urn:uuid:" + string(data[oId:oDate]) + ">"
	t := time.Time{}
	if err := t.UnmarshalBinary(data[oDate:oUri]); err != nil {
		return nil, err
	}
	r.TargetDate = t.Format(time.RFC3339)
	r.TargetUri = string(data[oUri:])
	return r, nil
}

func MarshalRevisitRef(r *gowarc.RevisitRef) (data []byte, err error) {
	id := strings.Trim(r.TargetRecordId, "<>")[9:]
	uri := r.TargetUri
	d, err := time.Parse(time.RFC3339, r.TargetDate)
	if err != nil {
		return nil, err
	}
	date, err := d.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var profile byte
	switch r.Profile {
	case gowarc.ProfileIdenticalPayloadDigestV1_0:
		profile = 1
	case gowarc.ProfileIdenticalPayloadDigestV1_1:
		profile = 2
	case gowarc.ProfileServerNotModifiedV1_0:
		profile = 3
	case gowarc.ProfileServerNotModifiedV1_1:
		profile = 4
	}

	length := oUri + len(uri)
	b := make([]byte, length)
	b[0] = profile
	copy(b[oId:], id)
	copy(b[oDate:], date)
	copy(b[oUri:], uri)
	return b, nil
}
