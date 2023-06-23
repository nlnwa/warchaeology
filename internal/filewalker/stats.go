package filewalker

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type Result interface {
	fmt.Stringer
	Log(fileNum int) string
	GetStats() Stats
	IncrRecords()
	IncrProcessed()
	IncrDuplicates()
	AddError(err error)
	Records() int64
	Processed() int64
	ErrorCount() int64
	Errors() []error
	SetFatal(error)
	Fatal() error
	Duplicates() int64
	error
}

type Stats interface {
	fmt.Stringer
	Merge(s Stats)
}

type stats struct {
	files      int
	records    int64
	processed  int64
	errors     int64
	duplicates int64
}

func NewStats() *stats {
	return &stats{}
}

func (s *stats) String() string {
	return fmt.Sprintf("files: %d, records: %d, processed: %d, errors: %d, duplicates: %d", s.files, s.records, s.processed, s.errors, s.duplicates)
}

func (s *stats) Merge(s2 Stats) {
	s.files++
	if stats, ok := s2.(*stats); ok {
		s.records += stats.records
		s.processed += stats.processed
		s.errors += stats.errors
		s.duplicates += stats.duplicates
	}
}

type result struct {
	fileName   string
	records    int64
	processed  int64
	errorCount int64
	errors     []error
	duplicates int64
	fatal      error
}

func NewResult(fileName string) Result {
	return &result{fileName: fileName}
}

func (r *result) IncrRecords() {
	r.records++
}

func (r *result) IncrProcessed() {
	r.processed++
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

func (r *result) Processed() int64 {
	return r.processed
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
	sb := strings.Builder{}
	if len(r.errors) > 0 {
		for i, e := range r.errors {
			if i > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(fmt.Sprintf("   %s", e))
		}
	}
	return sb.String()
}

func (r *result) SetFatal(e error) {
	r.fatal = e
}

func (r *result) Fatal() error {
	return r.fatal
}

func (r *result) String() string {
	return fmt.Sprintf("%s: records: %d, processed: %d, errors: %d, duplicates: %d", r.fileName, r.records, r.processed, r.ErrorCount(), r.duplicates)
}

func (r *result) Log(fileNum int) string {
	return fmt.Sprintf("%06d %s", fileNum, r.String())
}

func (r *result) GetStats() Stats {
	return &stats{
		records:    r.records,
		processed:  r.processed,
		errors:     r.ErrorCount(),
		duplicates: r.duplicates,
	}
}

func (r *result) UnmarshalBinary(data []byte) error {
	var read int
	v, n := binary.Varint(data[read:])
	r.records = v
	read += n
	v, n = binary.Varint(data[read:])
	r.processed = v
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
	written += binary.PutVarint(b, r.processed)
	b = buf[written:]
	written += binary.PutVarint(b, r.duplicates)
	b = buf[written:]
	written += binary.PutVarint(b, r.errorCount)

	return buf[:written], nil
}
