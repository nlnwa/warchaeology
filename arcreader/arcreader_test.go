package arcreader

import (
	"path/filepath"
	"testing"

	"github.com/nationallibraryofnorway/warchaeology/v4/internal/warc"
	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/whatwg-url/url"
	"github.com/spf13/afero"
)

var (
	testDataDir = filepath.Join("..", "testdata")
)

var testFiles = map[string]string{
	// from https://archive.org/details/SAMPLE_ARC_WHITEHOUSE
	"sample_arc_whitehouse": filepath.Join(testDataDir, "arc", "ARC-SAMPLE-20060928223931-00000-gojoblack.arc.gz"),
	"invalid-host":          filepath.Join(testDataDir, "arc", "invalid-host.arc"),
}

func TestArcReader(t *testing.T) {
	tests := []struct {
		name              string
		wantErr           bool
		warcRecordOptions []gowarc.WarcRecordOption
	}{
		{
			name: "sample_arc_whitehouse",
		},
		{
			name:    "invalid-host",
			wantErr: true,
		},
		{
			name: "invalid-host",
			warcRecordOptions: []gowarc.WarcRecordOption{
				gowarc.WithUrlParserOptions(url.WithLaxHostParsing()),
			},
		},
	}

	defaultWarcRecordOptions := []gowarc.WarcRecordOption{
		gowarc.WithBufferTmpDir(t.TempDir()),
		gowarc.WithStrictValidation(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// resolve test file path
			testFile, err := filepath.Abs(testFiles[test.name])
			if err != nil {
				t.Fatal(err)
			}

			warcRecordOptions := defaultWarcRecordOptions
			if test.warcRecordOptions != nil {
				warcRecordOptions = append(warcRecordOptions, test.warcRecordOptions...)
			}

			arcFileReader, err := NewArcFileReader(afero.NewReadOnlyFs(afero.NewOsFs()), testFile, 0, warcRecordOptions...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, err := range warc.Records(arcFileReader, nil, 0, 0) {
				if err != nil && !test.wantErr {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
