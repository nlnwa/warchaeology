package cat

import (
	"path/filepath"
	"testing"
)

func BenchmarkListRecords(b *testing.B) {
	testDataDir := filepath.Join("..", "..", "test-data")
	catConfig := &config{}
	fileName := filepath.Join(testDataDir, "samsung-with-error", "rec-33318048d933-20240317162652059-0.warc.gz")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		listRecords(catConfig, fileName)
	}
}
