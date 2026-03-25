package warc

import (
	"iter"
	"strings"

	"github.com/nationallibraryofnorway/warchaeology/v5/internal/filter"
	"github.com/nlnwa/gowarc/v3"
)

func Filter(seq iter.Seq2[gowarc.Record, error], accept func(gowarc.Record) bool) iter.Seq2[gowarc.Record, error] {
	if accept == nil {
		return seq
	}
	return func(yield func(gowarc.Record, error) bool) {
		for rec, err := range seq {
			if err != nil {
				yield(rec, err)
				return
			}
			if !accept(rec) {
				_ = rec.Close()
				continue
			}
			if !yield(rec, nil) {
				return
			}
		}
	}
}

func Limit(seq iter.Seq2[gowarc.Record, error], limit int) iter.Seq2[gowarc.Record, error] {
	if limit <= 0 {
		return seq
	}

	return func(yield func(gowarc.Record, error) bool) {
		count := 0
		for rec, err := range seq {
			if err != nil {
				yield(rec, err)
				return
			}
			count++
			if !yield(rec, nil) {
				return
			}
			if count >= limit {
				return
			}
		}
	}
}

func Nth(seq iter.Seq2[gowarc.Record, error], n int) iter.Seq2[gowarc.Record, error] {
	return func(yield func(gowarc.Record, error) bool) {
		if n <= 0 {
			return
		}

		count := 0
		for rec, err := range seq {
			if err != nil {
				yield(rec, err)
				return
			}
			count++
			if count == n {
				yield(rec, nil)
				return
			}
			_ = rec.Close()
		}
	}
}

func Skip(seq iter.Seq2[gowarc.Record, error], n int) iter.Seq2[gowarc.Record, error] {
	if n <= 0 {
		return seq
	}

	return func(yield func(gowarc.Record, error) bool) {
		skipped := 0
		for rec, err := range seq {
			if err != nil {
				yield(rec, err)
				return
			}
			if skipped < n {
				skipped++
				_ = rec.Close()
				continue
			}
			if !yield(rec, nil) {
				return
			}
		}
	}
}

func Compose(seq iter.Seq2[gowarc.Record, error], recordFilter *filter.RecordFilter, nth int, limit int) iter.Seq2[gowarc.Record, error] {
	if recordFilter != nil {
		seq = Filter(seq, ByRecordFilter(recordFilter))
	}
	if nth > 0 {
		return Nth(seq, nth)
	}
	if limit > 0 {
		return Limit(seq, limit)
	}
	return seq
}

func ByRecordFilter(recordFilter *filter.RecordFilter) func(gowarc.Record) bool {
	if recordFilter == nil {
		return func(gowarc.Record) bool { return true }
	}
	return func(record gowarc.Record) bool {
		if record.WarcRecord == nil {
			return false
		}
		return recordFilter.Accept(record.WarcRecord)
	}
}

func ByRecordType(types ...gowarc.RecordType) func(gowarc.Record) bool {
	var mask gowarc.RecordType
	for _, t := range types {
		mask |= t
	}

	return func(record gowarc.Record) bool {
		if record.WarcRecord == nil {
			return false
		}
		return record.WarcRecord.Type()&mask != 0
	}
}

func ByContentType(substr string) func(gowarc.Record) bool {
	substr = strings.ToLower(substr)
	return func(record gowarc.Record) bool {
		if record.WarcRecord == nil {
			return false
		}
		ct := strings.ToLower(record.WarcRecord.WarcHeader().Get(gowarc.ContentType))
		return strings.Contains(ct, substr)
	}
}

func ByTargetURI(substr string) func(gowarc.Record) bool {
	return func(record gowarc.Record) bool {
		if record.WarcRecord == nil {
			return false
		}
		uri := record.WarcRecord.WarcHeader().Get(gowarc.WarcTargetURI)
		return strings.Contains(uri, substr)
	}
}

func And(predicates ...func(gowarc.Record) bool) func(gowarc.Record) bool {
	return func(record gowarc.Record) bool {
		for _, predicate := range predicates {
			if !predicate(record) {
				return false
			}
		}
		return true
	}
}

func Or(predicates ...func(gowarc.Record) bool) func(gowarc.Record) bool {
	return func(record gowarc.Record) bool {
		for _, predicate := range predicates {
			if predicate(record) {
				return true
			}
		}
		return false
	}
}

func Not(predicate func(gowarc.Record) bool) func(gowarc.Record) bool {
	return func(record gowarc.Record) bool {
		return !predicate(record)
	}
}
