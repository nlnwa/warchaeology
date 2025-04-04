package nedlib

import (
	"path/filepath"
	"testing"

	"github.com/nlnwa/warchaeology/v4/internal/warcwriterconfig"
	"github.com/spf13/afero"
)

var (
	testDataDir = filepath.Join("..", "..", "..", "testdata")
)

func Test(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
	}{
		{
			name: filepath.Join(testDataDir, "nedlib", "nb-image", "b863a630196bce1a15ca86b40f34a2d5.meta"),
		},
		{
			name: filepath.Join(testDataDir, "nedlib", "nb-image", "e4a2d28bdf4c38b8f6f291f7c8c958d5.meta"),
		},
		{
			name:      filepath.Join(testDataDir, "nedlib", "bad", "df94a25cc254ba3c9765c5263b7ca6aa.meta"),
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
				warcwriterconfig.WithWarcFileNameGenerator("nedlib"),
				warcwriterconfig.WithCompress(false),
			)
			if err != nil {
				t.Fatalf("failed to create warc writer config: %v", err)
			}
			defer warcWriterConfig.Close()

			o := &ConvertNedlibOptions{
				WarcWriterConfig: warcWriterConfig,
			}

			_, err = o.handleFile(afero.NewOsFs(), filename)
			if err != nil && !tt.wantError {
				t.Fatal(err)
			}
		})
	}
}
