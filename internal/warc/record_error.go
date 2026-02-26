package warc

import (
	"github.com/nlnwa/gowarc/v3"
)

func ErrorFrom(record gowarc.Record, err error) RecordError {
	return RecordError{offset: record.Offset, Err: err}
}

type RecordError struct {
	offset int64
	Err    error
}

func (e RecordError) Unwrap() error {
	return e.Err
}

func (e RecordError) Error() string {
	return e.Err.Error()
}

func (e RecordError) Offset() int64 {
	return e.offset
}
