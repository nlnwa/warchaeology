package nedlib

import (
	"path/filepath"
	"testing"

	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/spf13/afero"
)

type walker []string

func (fileWalker *walker) dummyWalkFunction(_ afero.Fs, path string) filewalker.Result {
	*fileWalker = append(*fileWalker, path)
	return filewalker.NewResult(path)
}

func TestConvert(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "test-data")
	nedlibDir := filepath.Join(testDataDir, "nedlib", "nb-image")

	fileWalker := walker{}
	config := &conf{}
	config.files = []string{nedlibDir}
	_, err := filewalker.New([]string{nedlibDir}, false, false, []string{}, 1, fileWalker.dummyWalkFunction)
	if err != nil {
		t.Errorf("error creating file walker, original error: '%v'", err)
	}
}
