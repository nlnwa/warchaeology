package warc

import (
	"errors"
	"iter"
	"testing"
	"time"

	"github.com/nationallibraryofnorway/warchaeology/v5/internal/filter"
	"github.com/nlnwa/gowarc/v3"
)

func TestFilterLimitAndNth(t *testing.T) {
	seq := recordsSeq(t,
		newTestRecord(t, gowarc.Response, "https://example.org/1", "text/html"),
		newTestRecord(t, gowarc.Resource, "https://example.org/2", "image/png"),
		newTestRecord(t, gowarc.Response, "https://example.org/3", "text/plain"),
	)

	filtered := Filter(seq, ByRecordType(gowarc.Response))
	limited := Limit(filtered, 1)

	count := 0
	for rec, err := range limited {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		if rec.WarcRecord.Type() != gowarc.Response {
			t.Fatalf("expected response record, got %v", rec.WarcRecord.Type())
		}
		_ = rec.Close()
	}

	if count != 1 {
		t.Fatalf("expected 1 record, got %d", count)
	}

	nth := Nth(seq, 2)
	count = 0
	for rec, err := range nth {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		if got := rec.WarcRecord.WarcHeader().Get(gowarc.WarcTargetURI); got != "https://example.org/2" {
			t.Fatalf("expected second record URI, got %s", got)
		}
		_ = rec.Close()
	}
	if count != 1 {
		t.Fatalf("expected nth to return 1 record, got %d", count)
	}
}

func TestComposePreservesNthOverLimit(t *testing.T) {
	seq := recordsSeq(t,
		newTestRecord(t, gowarc.Response, "https://example.org/1", "text/html"),
		newTestRecord(t, gowarc.Response, "https://example.org/2", "text/html"),
		newTestRecord(t, gowarc.Response, "https://example.org/3", "text/html"),
	)

	recordFilter := filter.New(filter.WithRecordTypes(gowarc.Response))
	composed := Compose(seq, recordFilter, 2, 1)

	count := 0
	for rec, err := range composed {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		if got := rec.WarcRecord.WarcHeader().Get(gowarc.WarcTargetURI); got != "https://example.org/2" {
			t.Fatalf("expected URI https://example.org/2, got %s", got)
		}
		_ = rec.Close()
	}

	if count != 1 {
		t.Fatalf("expected 1 record, got %d", count)
	}
}

func TestFilterPassesThroughError(t *testing.T) {
	testErr := errors.New("boom")
	seq := func(yield func(gowarc.Record, error) bool) {
		rec := gowarc.Record{WarcRecord: newTestRecord(t, gowarc.Response, "https://example.org/1", "text/html")}
		if !yield(rec, nil) {
			return
		}
		yield(gowarc.Record{}, testErr)
	}

	count := 0
	filtered := Filter(seq, ByRecordType(gowarc.Response))
	for rec, err := range filtered {
		if err != nil {
			if !errors.Is(err, testErr) {
				t.Fatalf("expected %v, got %v", testErr, err)
			}
			break
		}
		count++
		_ = rec.Close()
	}

	if count != 1 {
		t.Fatalf("expected 1 record before error, got %d", count)
	}
}

func recordsSeq(t *testing.T, records ...gowarc.WarcRecord) iter.Seq2[gowarc.Record, error] {
	t.Helper()
	return func(yield func(gowarc.Record, error) bool) {
		for _, record := range records {
			if !yield(gowarc.Record{WarcRecord: record}, nil) {
				return
			}
		}
	}
}

func newTestRecord(t *testing.T, recordType gowarc.RecordType, targetURI string, contentType string) gowarc.WarcRecord {
	t.Helper()

	rb := gowarc.NewRecordBuilder(recordType)
	rb.AddWarcHeader(gowarc.WarcTargetURI, targetURI)
	rb.AddWarcHeader(gowarc.ContentType, contentType)
	rb.AddWarcHeaderTime(gowarc.WarcDate, time.Now().UTC())
	rb.AddWarcHeaderInt64(gowarc.ContentLength, 0)

	rec, _, err := rb.Build()
	if err != nil {
		t.Fatalf("failed to build record: %v", err)
	}

	return rec
}
