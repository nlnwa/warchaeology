package validate

import (
	"errors"
	"io"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

var testDataDir = filepath.Join("..", "..", "testdata")

func TestValidateFile(t *testing.T) {
	var tests = []struct {
		name string
		err  error
	}{
		{
			name: filepath.Join(testDataDir, "warc", "single-record.warc"),
		},
		{
			name: filepath.Join(testDataDir, "warc", "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz"),
			err:  io.ErrUnexpectedEOF,
		},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.name), func(t *testing.T) {
			options := &ValidateOptions{
				outputDir: t.TempDir(),
			}

			results, err := options.handleFile(afero.NewOsFs(), tt.name)
			if err != nil && !errors.Is(err, tt.err) {
				t.Fatalf("expected error %T: %v, got %T: %v", tt.err, tt.err, err, err)
			}

			if results.ErrorCount() > 0 {
				for _, err := range results.Errors() {
					t.Error(err)
				}
			}
		})
	}
}
