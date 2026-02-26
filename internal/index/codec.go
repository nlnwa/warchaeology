package index

import (
	"fmt"
	"strings"
	"time"

	"github.com/nlnwa/gowarc/v3"
)

const (
	oProfile = 0
	oId      = oProfile + 1
	oDate    = oId + 36
	oUri     = oDate + 15
)

func UnmarshalRevisitRef(data []byte) (*gowarc.RevisitRef, error) {
	if len(data) < oUri {
		return nil, fmt.Errorf("invalid revisit reference encoding: length %d is smaller than minimum %d", len(data), oUri)
	}

	revisitReference := &gowarc.RevisitRef{}
	switch data[0] {
	case 1:
		revisitReference.Profile = gowarc.ProfileIdenticalPayloadDigestV1_0
	case 2:
		revisitReference.Profile = gowarc.ProfileIdenticalPayloadDigestV1_1
	case 3:
		revisitReference.Profile = gowarc.ProfileServerNotModifiedV1_0
	case 4:
		revisitReference.Profile = gowarc.ProfileServerNotModifiedV1_1
	default:
		return nil, fmt.Errorf("invalid revisit profile encoding: %d", data[0])
	}

	revisitReference.TargetRecordId = "<urn:uuid:" + string(data[oId:oDate]) + ">"
	now := time.Time{}
	if err := now.UnmarshalBinary(data[oDate:oUri]); err != nil {
		return nil, err
	}
	revisitReference.TargetDate = now.Format(time.RFC3339)
	revisitReference.TargetUri = string(data[oUri:])
	return revisitReference, nil
}

func MarshalRevisitRef(revisitReference *gowarc.RevisitRef) (data []byte, err error) {
	if revisitReference == nil {
		return nil, fmt.Errorf("revisit reference is nil")
	}

	id, err := normalizeRecordID(revisitReference.TargetRecordId)
	if err != nil {
		return nil, err
	}
	uri := revisitReference.TargetUri
	time, err := time.Parse(time.RFC3339, revisitReference.TargetDate)
	if err != nil {
		return nil, err
	}
	date, err := time.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var profile byte
	switch revisitReference.Profile {
	case gowarc.ProfileIdenticalPayloadDigestV1_0:
		profile = 1
	case gowarc.ProfileIdenticalPayloadDigestV1_1:
		profile = 2
	case gowarc.ProfileServerNotModifiedV1_0:
		profile = 3
	case gowarc.ProfileServerNotModifiedV1_1:
		profile = 4
	default:
		return nil, fmt.Errorf("invalid revisit profile: %q", revisitReference.Profile)
	}

	length := oUri + len(uri)
	b := make([]byte, length)
	b[0] = profile
	copy(b[oId:], id)
	copy(b[oDate:], date)
	copy(b[oUri:], uri)
	return b, nil
}

func normalizeRecordID(recordID string) (string, error) {
	trimmed := strings.Trim(recordID, "<>")
	if !strings.HasPrefix(trimmed, "urn:uuid:") {
		return "", fmt.Errorf("invalid target record id: %q", recordID)
	}
	id := strings.TrimPrefix(trimmed, "urn:uuid:")
	if len(id) != 36 {
		return "", fmt.Errorf("invalid target record id length: %d", len(id))
	}
	return id, nil
}
