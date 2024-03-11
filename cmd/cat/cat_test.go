package cat

import (
	"testing"
	"time"
)

func BenchmarkDummy(b *testing.B) {
	// This is a dummy test, it should be replaced with something more
	// meaningful in a later commit
	for i := 0; i < b.N; i++ {
		time.Sleep(1 * time.Nanosecond)
	}
}
