package validate

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestValidateSamsungFileWithError(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "testdata")
	warcWithErrors := filepath.Join(testDataDir, "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz")
	_ = validateFile(afero.NewOsFs(), warcWithErrors)
}
