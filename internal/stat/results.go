package stat

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"strings"
)

type Result interface {
	fmt.Stringer
	error
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	Log(fileNum int) string
	Name() string
	IncrRecords()
	IncrDuplicates()
	AddError(err error)
	Records() int64
	ErrorCount() int64
	Errors() []error
	Duplicates() int64
	SetHash(hash string)
	Hash() string
}

type result struct {
	fileName   string
	records    int64
	errorCount int64
	errors     []error
	duplicates int64
	hash       string
}

func NewResult(fileName string) Result {
	return &result{fileName: fileName}
}

func (r *result) Name() string {
	return r.fileName
}

func (r *result) IsValid() bool {
	return r.errorCount == 0
}

func (r *result) IncrRecords() {
	r.records++
}

func (r *result) IncrDuplicates() {
	r.duplicates++
}

func (r *result) AddError(err error) {
	r.errors = append(r.errors, err)
	r.errorCount++
}

func (r *result) Records() int64 {
	return r.records
}

func (r *result) ErrorCount() int64 {
	return r.errorCount
}

func (r *result) Duplicates() int64 {
	return r.duplicates
}

func (r *result) Errors() []error {
	return r.errors
}

func (r *result) Error() string {
	if len(r.errors) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for i, e := range r.errors {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(fmt.Sprintf("   %s", e))
	}
	return sb.String()
}

func (r *result) String() string {
	return fmt.Sprintf("%s: records: %d, errors: %d, duplicates: %d", r.fileName, r.records, r.ErrorCount(), r.duplicates)
}

func (r *result) Log(fileNum int) string {
	return fmt.Sprintf("%06d %s", fileNum, r.String())
}

func (r *result) SetHash(hash string) {
	r.hash = hash
}

func (r *result) Hash() string {
	return r.hash
}

func (r *result) UnmarshalBinary(data []byte) error {
	var read int
	v, n := binary.Varint(data[read:])
	r.records = v
	read += n
	v, n = binary.Varint(data[read:])
	r.duplicates = v
	read += n
	v, _ = binary.Varint(data[read:])
	r.errorCount = v

	return nil
}

func (r *result) MarshalBinary() (data []byte, err error) {
	buf := make([]byte, binary.MaxVarintLen64*3)
	written := binary.PutVarint(buf, r.records)
	b := buf[written:]
	written += binary.PutVarint(b, r.duplicates)
	b = buf[written:]
	written += binary.PutVarint(b, r.errorCount)

	return buf[:written], nil
}
