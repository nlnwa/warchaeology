package arcreader

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v3/internal/warc"
	"github.com/spf13/afero"
)

var (
	testDataDir = filepath.Join("..", "testdata")
)

var testFiles = map[string]string{
	// from https://archive.org/details/SAMPLE_ARC_WHITEHOUSE
	"sample_arc_whitehouse": filepath.Join(testDataDir, "arc", "ARC-SAMPLE-20060928223931-00000-gojoblack.arc.gz"),
}

func TestArcReader(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "sample_arc_whitehouse",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// resolve test file path
			testFile, err := filepath.Abs(testFiles[test.name])
			if err != nil {
				t.Fatal(err)
			}

			arcFileReader, err := NewArcFileReader(afero.NewReadOnlyFs(afero.NewOsFs()), testFile, 0,
				gowarc.WithBufferTmpDir(t.TempDir()),
				gowarc.WithStrictValidation(),
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for record := range warc.NewIterator(context.Background(), arcFileReader, nil, 0, 0) {
				if record.Err != nil {
					t.Errorf("unexpected error: %v", record.Err)
				}
			}
		})
	}
}
