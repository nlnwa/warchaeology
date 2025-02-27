package warc

import (
	"errors"
	"io"
	"path/filepath"
	"testing"

	"github.com/nlnwa/gowarc/v2"
)

var (
	testDataDir = filepath.Join("..", "..", "testdata")
)

var testFiles = map[string]string{
	"empty":              filepath.Join(testDataDir, "warc", "empty.warc"),
	"single-record":      filepath.Join(testDataDir, "warc", "single-record.warc"),
	"samsung-with-error": filepath.Join(testDataDir, "warc", "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz"),
}

var tests = []struct {
	file         string // path to WARC file
	limit        int    // limit number of recor
	nth          int    // return only the Nth record
	wantCount    int    // expected number of records from iterator
	wantRecordId string // expected record id
	err          error  // expected error
}{
	{
		file:      testFiles["empty"],
		wantCount: 0,
	},
	{
		file:      testFiles["single-record"],
		wantCount: 1,
	},
	{
		file:      testFiles["samsung-with-error"],
		wantCount: 53,
		err:       io.ErrUnexpectedEOF,
	},
	{
		file:      testFiles["samsung-with-error"],
		limit:     50,
		wantCount: 50,
	},
	{
		file:         testFiles["samsung-with-error"],
		nth:          7,
		wantCount:    1,
		wantRecordId: "urn:uuid:60331c1b-c2f4-486a-a14f-bd448ba6e1c7",
	},
}

func TestIterator(t *testing.T) {
	for _, test := range tests {
		t.Run(filepath.Base(test.file), func(t *testing.T) {
			// capture test variable from outer scope
			test := test

			// resolve test file path
			testFile, err := filepath.Abs(test.file)
			if err != nil {
				t.Fatal(err)
			}

			// open WARC file reader
			warcFileReader, err := gowarc.NewWarcFileReader(testFile, 0)
			if err != nil {
				t.Fatal(err)
			}

			// count records
			count := 0

			// iterate over the records channel
			for record, err := range Records(warcFileReader, nil, test.nth, test.limit) {
				if err != nil {
					if !errors.Is(err, test.err) {
						t.Fatal(err)
					}
					break
				}
				count++

				if test.wantRecordId != "" {
					recordId := record.WarcRecord.WarcHeader().GetId(gowarc.WarcRecordID)
					// assert record id
					if recordId != test.wantRecordId {
						t.Errorf("expected record id %s, got %s", test.wantRecordId, recordId)
					}
				}
			}

			// assert number of records from iterator
			if count != test.wantCount {
				t.Errorf("expected %d records, got %d", test.wantCount, count)
			}
		})
	}
}
