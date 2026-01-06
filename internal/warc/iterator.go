package warc

import (
	"io"
	"iter"

	"github.com/nationallibraryofnorway/warchaeology/v4/internal/filter"
	"github.com/nlnwa/gowarc/v2"
)

// Record represents a WARC record with additional metadata such as offset,
// size, validation and error.
type Record struct {
	Offset     int64
	Size       int64
	WarcRecord gowarc.WarcRecord
	Validation *gowarc.Validation
}

func (r Record) Close() error {
	if r.WarcRecord != nil {
		return r.WarcRecord.Close()
	}
	return nil
}

type RecordIterator interface {
	Next() (gowarc.WarcRecord, int64, *gowarc.Validation, error)
}

// Records is an iterator over records that wraps a RecordIterator to
// filter and limit the records returned from the RecordIterator. Each
// gowarc.WarcRecord returned from the RecordIterator is wrapped in a Record
// type that provides metadata like offset, size, validation and error.
func Records(iter RecordIterator, filter *filter.RecordFilter, nth int, limit int) iter.Seq2[Record, error] {
	// Ignore limit if nth is specified
	if nth > 0 && limit > 0 {
		limit = 0
	}
	return func(yield func(Record, error) bool) {
		var currentOffset, previousOffset int64
		var warcRecord, previousWarcRecord gowarc.WarcRecord
		var validation, previousValidation *gowarc.Validation
		var err, previousErr error

		var skippedFirst bool

		var count int

		for {
			// return if previous record was the last record
			if err == io.EOF {
				return
			}

			// keep track of the previous iteration's values
			previousWarcRecord = warcRecord
			previousValidation = validation
			previousErr = err
			previousOffset = currentOffset

			// read next record
			warcRecord, currentOffset, validation, err = iter.Next()

			if !skippedFirst {
				// Delay sending the first record to be able to calculate the
				// record size which is the difference between the current and
				// the previous offset.
				skippedFirst = true
				continue
			}

			record := Record{
				WarcRecord: previousWarcRecord,
				Size:       currentOffset - previousOffset,
				Offset:     previousOffset,
				Validation: previousValidation,
			}
			if previousErr != nil {
				if !yield(record, previousErr) {
					return
				} else {
					continue
				}
			}

			if filter != nil && !filter.Accept(record.WarcRecord) {
				continue
			}

			// keep track of the number of records accepted by the filter
			count++

			// if there was a Nth record specified, skip until we reach it
			if nth > 0 && count != nth {
				continue
			}
			if !yield(record, nil) {
				return
			}
			// If there was a Nth record specified, return after Nth record is reached
			if nth > 0 {
				return
			}
			if limit > 0 && count >= limit {
				return
			}
		}
	}
}

func Error(record Record, err error) RecordError {
	return RecordError{Record: record, Err: err}
}

type RecordError struct {
	Record Record
	Err    error
}

func (e RecordError) Unwrap() error {
	return e.Err
}

func (e RecordError) Error() string {
	return e.Err.Error()
}

func (e RecordError) Offset() int64 {
	return e.Record.Offset
}
