package validate

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestValidateSamsungFileWithError(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "test-data")
	warcWithErrors := filepath.Join(testDataDir, "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz")
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("validateFile did not panic")
		}
	}()
	_ = validateFile(afero.NewOsFs(), warcWithErrors)
}
