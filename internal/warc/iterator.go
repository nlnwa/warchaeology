package warc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nlnwa/gowarc"
	"github.com/nlnwa/warchaeology/internal/filter"
)

// Record represents a WARC record with additional metadata
type Record struct {
	Offset     int64
	Size       int64
	Err        error
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

// Itetaror is a WARC record iterator
type Iterator struct {
	// reader to read WARC records from
	WarcFileReader RecordIterator

	// return only the Nth record (0 for all) after applying filter
	Nth int

	// return at most N records (0 for all) after applying filter
	Limit int

	// return only records that match the filter
	Filter *filter.RecordFilter

	// channel to send records to
	Records chan<- Record
}

func NewIterator(ctx context.Context, reader RecordIterator, filter *filter.RecordFilter, nth, limit int) <-chan Record {
	records := make(chan Record)
	iterator := Iterator{
		WarcFileReader: reader,
		Filter:         filter,
		Nth:            nth,
		Limit:          limit,
		Records:        records,
	}
	go iterator.iterate(ctx)

	return records
}

// Iterate reads WARC records from the WARC file reader and sends them to the result channel
func (iter Iterator) iterate(ctx context.Context) {
	defer close(iter.Records)

	var currentOffset, previousOffset int64
	var warcRecord, previousWarcRecord gowarc.WarcRecord
	var validation, previousValidation *gowarc.Validation
	var err, previousErr error

	var skippedFirst bool

	var count int

	for {
		// check if context is done
		select {
		case <-ctx.Done():
			return
		default:
		}
		// return if previous record was the last record
		if errors.Is(err, io.EOF) {
			return
		}

		// keep track of the previous iteration's values
		previousWarcRecord = warcRecord
		previousValidation = validation
		previousErr = err
		previousOffset = currentOffset

		// read next record
		warcRecord, currentOffset, validation, err = iter.WarcFileReader.Next()

		if !skippedFirst {
			// we delay sending the first record to be able to calculate record size
			skippedFirst = true
			continue
		}

		record := Record{
			WarcRecord: previousWarcRecord,
			Size:       currentOffset - previousOffset,
			Offset:     previousOffset,
			Validation: previousValidation,
			Err:        previousErr,
		}

		// if record has an error, send it and return
		if record.Err != nil {
			select {
			case <-ctx.Done():
				return
			case iter.Records <- record:
			}
			return
		}

		if iter.Filter != nil && !iter.Filter.Accept(record.WarcRecord) {
			continue
		}

		// keep track of the number of records accepted by the filter
		count++

		// if there was a Nth record specified, skip until we reach it
		if iter.Nth > 0 && count != iter.Nth {
			continue
		}

		select {
		case <-ctx.Done():
			return
		case iter.Records <- record:
			// If there was a Nth record specified, return after sending it
			if iter.Nth > 0 {
				return
			}
			// If there was a limit specified, return after limit is reached
			if iter.Limit > 0 && count >= iter.Limit {
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

func (e RecordError) Error() string {
	return fmt.Sprintf("offset: %d: %v", e.Record.Offset, e.Err)
}
