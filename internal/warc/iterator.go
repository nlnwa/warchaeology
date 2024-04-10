package warc

import (
	"context"
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

// Itetaror is a WARC record iterator
type Iterator struct {
	// reader to read WARC records from
	WarcFileReader *gowarc.WarcFileReader

	// return only the Nth record (0 for all) after applying filter
	Nth int

	// return at most N records (0 for all) after applying filter
	Limit int

	// return only records that match the filter
	Filter *filter.Filter

	// channel to send records to
	Records chan<- Record
}

// Iterate reads WARC records from the WARC file reader and sends them to the result channel
func (iterator Iterator) Iterate(ctx context.Context) {
	// Assert that the required fields are set
	if iterator.WarcFileReader == nil {
		panic("WarcFileReader is nil")
	}
	if iterator.Records == nil {
		panic("Result channel is nil")
	}

	defer close(iterator.Records)

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
		warcRecord, currentOffset, validation, err = iterator.WarcFileReader.Next()

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

		if iterator.Filter != nil && !iterator.Filter.Accept(record.WarcRecord) {
			continue
		}

		// keep track of the number of records accepted by the filter
		count++

		// if there was a Nth record specified, skip until we reach it
		if iterator.Nth > 0 && count != iterator.Nth {
			continue
		}

		select {
		case <-ctx.Done():
			return
		case iterator.Records <- record:
			// If there was a Nth record specified, return after sending it
			if iterator.Nth > 0 {
				return
			}
			// If there was a limit specified, return after limit is reached
			if iterator.Limit > 0 && count >= iterator.Limit {
				return
			}
		}
	}
}
