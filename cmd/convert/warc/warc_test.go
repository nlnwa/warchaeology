package warc

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/nlnwa/warchaeology/v3/internal/warcwriterconfig"
	"github.com/spf13/afero"
)

var (
	testDataDir = filepath.Join("..", "..", "..", "testdata")
)

var testFiles = map[string]string{
	"empty":              filepath.Join(testDataDir, "warc", "empty.warc"),
	"single-record":      filepath.Join(testDataDir, "warc", "single-record.warc"),
	"samsung-with-error": filepath.Join(testDataDir, "warc", "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz"),
}

func TestConvertWarcFile(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
	}{
		{
			name: testFiles["empty"],
		},
		{
			name: testFiles["single-record"],
		},
		{
			name:      testFiles["samsung-with-error"],
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename, err := filepath.Abs(tt.name)
			if err != nil {
				t.Fatalf("failed to get absolute path: %v", err)
			}

			warcWriterConfig, err := warcwriterconfig.New("test",
				warcwriterconfig.WithBufferTmpDir(t.TempDir()),
				warcwriterconfig.WithOutDir(t.TempDir()),
				warcwriterconfig.WithCompress(false),
			)
			if err != nil {
				t.Fatalf("failed to create warc writer config: %v", err)
			}
			defer warcWriterConfig.Close()

			o := &ConvertWarcOptions{
				WarcWriterConfig: warcWriterConfig,
			}

			result := o.readFile(context.Background(), afero.NewOsFs(), filename)
			if !tt.wantError && result.ErrorCount() > 0 {
				for _, err := range result.Errors() {
					t.Errorf("readFile() error = %v", err)
				}
				if result.Fatal() != nil {
					t.Errorf("readFile() fatal = %v", result.Fatal())
				}
			} else if tt.wantError && result.ErrorCount() == 0 {
				t.Errorf("readFile() expected error, got none")
			}
		})
	}
}
