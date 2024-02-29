package cat

import (
	"testing"
	"time"

	"github.com/nlnwa/warchaeology/internal/filter"
)

func BenchmarkDummy(b *testing.B) {
	// This is a dummy test, it should be replaced with something more
	// meaningful in a later commit
	for i := 0; i < b.N; i++ {
		time.Sleep(1 * time.Nanosecond)
	}
}

func testDir() string {
	return "../../test-data"
}

func TestReadValidWarc(t *testing.T) {
	config := &config{
		fileName: testDir() + "/wikipedia/labrador_retriever_0.warc.gz"}
	config.filter = filter.NewFromViper()
	listRecords(config, config.fileName)
}
