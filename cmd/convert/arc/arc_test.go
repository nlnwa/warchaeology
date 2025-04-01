package arc

import (
	"path/filepath"
	"testing"

	"github.com/nlnwa/warchaeology/v4/internal/warcwriterconfig"
	"github.com/spf13/afero"
)

var (
	testDataDir = filepath.Join("..", "..", "..", "testdata")
)

func TestConvertArcFile(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
	}{
		{
			name: filepath.Join(testDataDir, "arc", "ARC-SAMPLE-20060928223931-00000-gojoblack.arc.gz"),
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

			o := &ConvertArcOptions{
				WarcWriterConfig: warcWriterConfig,
			}

			result, err := o.handleFile(afero.NewOsFs(), filename)
			if err != nil {
				if !tt.wantError {
					t.Fatal(err)
				}
			}
			if !tt.wantError && result.ErrorCount() > 0 {
				for _, err := range result.Errors() {
					t.Error(err)
				}
			} else if tt.wantError && result.ErrorCount() == 0 {
				t.Errorf("readFile() expected error, got none")
			}
		})
	}
}
