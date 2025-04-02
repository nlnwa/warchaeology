package warc

import (
	"path/filepath"
	"testing"

	"github.com/nlnwa/gowarc/v2"
	"github.com/nlnwa/warchaeology/v4/internal/warc"
	"github.com/nlnwa/warchaeology/v4/internal/warcwriterconfig"
	"github.com/spf13/afero"
)

var (
	testDataDir = filepath.Join("..", "..", "..", "testdata")
)

var testFiles = map[string]string{
	"empty":              filepath.Join(testDataDir, "warc", "empty.warc"),
	"single-record":      filepath.Join(testDataDir, "warc", "single-record.warc"),
	"samsung-with-error": filepath.Join(testDataDir, "warc", "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz"),
	"convert":            filepath.Join(testDataDir, "warc", "convert.warc"),
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

			result, err := o.handleFile(afero.NewOsFs(), filename)
			if err != nil {
				if !tt.wantError {
					t.Fatal(err)
				}
			}
			if result.ErrorCount() > 0 {
				for _, err := range result.Errors() {
					t.Error(err)
				}
			}
		})
	}
}

func TestRepairWarcFile(t *testing.T) {
	// TODO add tests for every type of error that can be repaired and some that cannot
	tests := []struct {
		name         string
		wantNrErrors int64
	}{
		{
			name:         testFiles["convert"],
			wantNrErrors: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename, err := filepath.Abs(tt.name)
			if err != nil {
				t.Fatal(err)
			}

			outDir := t.TempDir()

			// Create a WARC writer config
			warcWriterConfig, err := warcwriterconfig.New("test",
				warcwriterconfig.WithBufferTmpDir(t.TempDir()),
				warcwriterconfig.WithOutDir(outDir),
				warcwriterconfig.WithCompress(false),
				warcwriterconfig.WithOneToOneWriter(true),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer warcWriterConfig.Close()

			// Set the WARC record options
			policy := gowarc.ErrWarn
			repair := true
			warcRecordOptions := []gowarc.WarcRecordOption{
				gowarc.WithSyntaxErrorPolicy(policy),
				gowarc.WithSpecViolationPolicy(policy),
				gowarc.WithAddMissingDigest(repair),
				gowarc.WithFixSyntaxErrors(repair),
				gowarc.WithFixDigest(repair),
				gowarc.WithAddMissingContentLength(repair),
				gowarc.WithAddMissingRecordId(repair),
				gowarc.WithFixContentLength(repair),
				gowarc.WithFixWarcFieldsBlockErrors(repair),
			}

			o := &ConvertWarcOptions{
				WarcWriterConfig:  warcWriterConfig,
				WarcRecordOptions: warcRecordOptions,
			}

			// Convert the WARC file
			result, err := o.handleFile(afero.NewOsFs(), filename)
			if err != nil {
				t.Fatal(err)
			}
			for _, err := range result.Errors() {
				t.Logf("Found error during conversion: %v", err)
			}
			if result.ErrorCount() != tt.wantNrErrors {
				t.Errorf("Expected %d errors, got %d", tt.wantNrErrors, result.ErrorCount())
			}

			// Validate the repaired WARC file
			outFile := filepath.Join(outDir, filepath.Base(filename))
			policy = gowarc.ErrWarn
			repair = false
			wr, err := gowarc.NewWarcFileReader(outFile, 0,
				gowarc.WithSyntaxErrorPolicy(policy),
				gowarc.WithSpecViolationPolicy(policy),
				gowarc.WithAddMissingDigest(repair),
				gowarc.WithFixSyntaxErrors(repair),
				gowarc.WithFixDigest(repair),
				gowarc.WithAddMissingContentLength(repair),
				gowarc.WithAddMissingRecordId(repair),
				gowarc.WithFixContentLength(repair),
				gowarc.WithFixWarcFieldsBlockErrors(repair),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = wr.Close() }()

			for record, err := range warc.Records(wr, nil, 0, 0) {
				if err != nil {
					t.Fatal(err)
				}
				for _, err = range *record.Validation {
					if err != nil {
						t.Error(err)
					}
				}
				record.Close()
			}
		})
	}
}
