package validate

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

var testDataDir = filepath.Join("..", "..", "testdata")

func TestValidateFile(t *testing.T) {
	var tests = []struct {
		name      string
		wantError bool
	}{
		{
			name: filepath.Join(testDataDir, "warc", "single-record.warc"),
		},
		{
			name:      filepath.Join(testDataDir, "warc", "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.name), func(t *testing.T) {
			options := &ValidateOptions{
				outputDir: t.TempDir(),
			}

			results := options.validateFile(context.TODO(), afero.NewOsFs(), tt.name)
			if !tt.wantError && results.ErrorCount() > 0 {
				for _, err := range results.Errors() {
					t.Errorf("validateFile() error = %v", err)
				}
				if results.Fatal() != nil {
					t.Errorf("validateFile() fatal error = %v", results.Fatal())
				}
			}
			if tt.wantError && results.ErrorCount() == 0 {
				t.Errorf("validateFile() expected error, but got none")
			}
		})
	}
}
