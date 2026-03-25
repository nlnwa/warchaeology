package index

import (
	"testing"

	"github.com/nationallibraryofnorway/warchaeology/v5/internal/stat"
)

func TestFileIndex_SaveAndGetFileStats(t *testing.T) {
	idx, err := NewFileIndex(t.TempDir(), true, true)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	result := stat.NewResult("a.warc")
	result.IncrRecords()
	result.IncrDuplicates()
	result.AddError(testError("boom"))

	if err := idx.SaveFileStats("a.warc", result); err != nil {
		t.Fatal(err)
	}

	got, err := idx.GetFileStats("a.warc")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.Records() != 1 || got.Duplicates() != 1 || got.ErrorCount() != 1 {
		t.Fatalf("unexpected stats: records=%d duplicates=%d errors=%d", got.Records(), got.Duplicates(), got.ErrorCount())
	}
}

func TestFileIndex_GetMissingFileStats(t *testing.T) {
	idx, err := NewFileIndex(t.TempDir(), true, true)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	got, err := idx.GetFileStats("missing.warc")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil result for missing key")
	}
}

type testError string

func (e testError) Error() string { return string(e) }
