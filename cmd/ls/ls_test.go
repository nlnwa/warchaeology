package ls

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nlnwa/warchaeology/internal/filter"
	"github.com/spf13/afero"
)

func TestConfigReadFileWithError(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "testdata")
	warcWithErrors := filepath.Join(testDataDir, "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz")
	config := &conf{}
	config.filter = filter.NewFromViper()
	config.files = []string{warcWithErrors}
	_ = config.readFile(afero.NewOsFs(), warcWithErrors)
	// TODO: check that the result contains the expected values
}

func BenchmarkReadFileWithError(b *testing.B) {
	// Stdout is redirected to /dev/null since the benchmarking tool
	// `github-action-benchmark` is unable to handle some of the output in the
	// benchmark result files.
	os.Stdout = nil
	testDataDir := filepath.Join("..", "..", "testdata")
	warcWithErrors := filepath.Join(testDataDir, "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz")
	config := &conf{}
	config.filter = filter.NewFromViper()
	config.files = []string{warcWithErrors}
	for i := 0; i < b.N; i++ {
		_ = config.readFile(afero.NewOsFs(), warcWithErrors)
	}
}
