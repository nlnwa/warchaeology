package stat

import (
	"fmt"
)

type Stats struct {
	Files      int
	Records    int64
	Errors     int64
	Duplicates int64
}

func NewStats() *Stats {
	return &Stats{}
}

func (s *Stats) String() string {
	return fmt.Sprintf("files: %d, records: %d, errors: %d, duplicates: %d", s.Files, s.Records, s.Errors, s.Duplicates)
}

func (s *Stats) Merge(result Result) {
	s.Files++
	s.Records += result.Records()
	s.Errors += result.ErrorCount()
	s.Duplicates += result.Duplicates()
}
