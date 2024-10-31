package cat

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v3/internal/warc"
)

var (
	testDataDir = filepath.Join("..", "..", "testdata")
)

var testFiles = map[string]string{
	"empty":              filepath.Join(testDataDir, "warc", "empty.warc"),
	"single-record":      filepath.Join(testDataDir, "warc", "single-record.warc"),
	"samsung-with-error": filepath.Join(testDataDir, "warc", "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz"),
}

// TestWriteWarcRecord tests the writeWarcRecord function.
func TestWriteWarcRecord(t *testing.T) {
	for name, testFile := range testFiles {
		t.Run(name, func(t *testing.T) {
			// capture testFile variable from outer scope
			testFile := testFile

			// resolve test file path
			testFile, err := filepath.Abs(testFile)
			if err != nil {
				t.Fatal(err)
			}

			var want []byte
			var warcFileReader *gowarc.WarcFileReader

			if filepath.Ext(testFile) == ".gz" {
				f, err := os.Open(testFile)
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()

				gzipReader, err := gzip.NewReader(f)
				if err != nil {
					t.Fatal(err)
				}
				defer gzipReader.Close()
				count := 0
				for {
					// Read bytes from the bufio reader
					buffer := make([]byte, 1024)
					n, err := gzipReader.Read(buffer)
					if err != nil {
						if err == io.EOF {
							break
						} else {
							want = append(want, buffer[:n]...)
							break
						}
					}
					count += n

					want = append(want, buffer[:n]...)
				}

				_, err = f.Seek(0, 0)
				if err != nil {
					t.Fatal(err)
				}
				err = gzipReader.Reset(f)
				if err != nil {
					t.Fatal(err)
				}

				// read uncompressed WARC file to get uncompressed offset
				warcFileReader, err = gowarc.NewWarcFileReaderFromStream(gzipReader, 0, gowarc.WithAddMissingDigest(false))
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					_ = warcFileReader.Close()
				}()
			} else {
				// open WARC file reader
				warcFileReader, err = gowarc.NewWarcFileReader(testFile, 0, gowarc.WithAddMissingDigest(false))
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					_ = warcFileReader.Close()
				}()

				want, err = os.ReadFile(testFile)
				if err != nil {
					t.Fatal(err)
				}
			}

			// print everything
			catWriter := &writer{
				showWarcHeader:     true,
				showProtocolHeader: true,
				showPayload:        true,
			}

			got := new(bytes.Buffer)
			var currentOffset int64

			for record := range warc.NewIterator(context.Background(), warcFileReader, nil, 0, 0) {
				if record.Err != nil {
					break
				}

				err := catWriter.writeWarcRecord(got, record.WarcRecord)
				if err != nil {
					t.Errorf("writeWarcRecord() error = %v", err)
				}

				currentOffset = record.Offset + record.Size
			}

			want = want[:currentOffset]

			var n = 10
			for math.Min(float64(len(got.Bytes())), float64(len(want))) < float64(n) {
				n--
			}

			if !bytes.Equal(got.Bytes(), want) {
				t.Errorf(`writeWarcRecord() = want != got
want is %d bytes, got is %d bytes
	-- first bytes of want: --
%v
	-- first bytes of got: --
%v
	-- last bytes of want: --
%v
	-- last bytes of got: --
%v
`,
					len(want), len(got.Bytes()),
					want[:n],
					got.Bytes()[:n],
					want[len(want)-n:],
					got.Bytes()[len(got.Bytes())-n:])
			}
		})
	}
}

func BenchmarkDummy(b *testing.B) {
	// This is a dummy test, it should be replaced with something more
	// meaningful in a later commit
	for i := 0; i < b.N; i++ {
		time.Sleep(1 * time.Nanosecond)
	}
}
