package filewalker

import (
	"fmt"
	"strings"
)

type Result interface {
	fmt.Stringer
	Log(fileNum int) string
	GetStats() Stats
	IncrRecords()
	IncrProcessed()
	AddError(err error)
	Records() int64
	Processed() int64
	ErrorCount() int64
	Errors() []error
	error
}

type Stats interface {
	fmt.Stringer
	Merge(s Stats)
}

type stats struct {
	files     int
	records   int64
	processed int64
	errors    int64
}

func NewStats() *stats {
	return &stats{}
}

func (s *stats) String() string {
	return fmt.Sprintf("files: %d, records: %d, processed: %d, errors: %d", s.files, s.records, s.processed, s.errors)
}

func (s *stats) Merge(s2 Stats) {
	s.files++
	if stats, ok := s2.(*stats); ok {
		s.records += stats.records
		s.processed += stats.processed
		s.errors += stats.errors
	}
}

type result struct {
	fileName  string
	records   int64
	processed int64
	errors    []error
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

func (r *result) AddError(err error) {
	r.errors = append(r.errors, err)
}

func (r *result) Records() int64 {
	return r.records
}

func (r *result) Processed() int64 {
	return r.processed
}

func (r *result) ErrorCount() int64 {
	return int64(len(r.errors))
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

func (r *result) String() string {
	return fmt.Sprintf("%s: records: %d, processed: %d, errors: %d", r.fileName, r.records, r.processed, r.ErrorCount())
}

func (r *result) Log(fileNum int) string {
	return fmt.Sprintf("%06d %s", fileNum, r.String())
}

func (r *result) GetStats() Stats {
	return &stats{
		records:   r.records,
		processed: r.processed,
		errors:    r.ErrorCount(),
	}
}
