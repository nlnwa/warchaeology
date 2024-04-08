package filewalker

import (
	"fmt"
)

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
