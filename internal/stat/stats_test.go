package stat

import "testing"

func TestCreateStats(t *testing.T) {
	stats := NewStats()
	if stats == nil {
		t.Errorf("Expected stats to be created")
	}
}

func TestStatsString(t *testing.T) {
	stats := NewStats()
	str := stats.String()
	if str == "" {
		t.Errorf("Expected stats string to be non-empty")
	}
}

func TestStatsFilesMerge(t *testing.T) {
	stats := NewStats()
	stats.files = 5
	stats2 := NewStats()
	stats2.files = 10
	stats.Merge(stats2)
	if stats.files != 6 {
		t.Errorf("Expected files to be 3, got %d", stats.files)
	}
	if stats2.files != 10 {
		t.Errorf("Expected files to be 2, got %d", stats2.files)
	}
}

func TestStatsRecordsMerge(t *testing.T) {
	stats := NewStats()
	stats.records = 5
	stats2 := NewStats()
	stats2.records = 10
	stats.Merge(stats2)
	if stats.records != 15 {
		t.Errorf("Expected records to be 15, got %d", stats.records)
	}
	if stats2.records != 10 {
		t.Errorf("Expected records to be 10, got %d", stats2.records)
	}
}

func TestStatsProcessedMerge(t *testing.T) {
	stats := NewStats()
	stats.processed = 5
	stats2 := NewStats()
	stats2.processed = 10
	stats.Merge(stats2)
	if stats.processed != 15 {
		t.Errorf("Expected processed to be 15, got %d", stats.processed)
	}
	if stats2.processed != 10 {
		t.Errorf("Expected processed to be 10, got %d", stats2.processed)
	}
}

func TestStatsErrorsMerge(t *testing.T) {
	stats := NewStats()
	stats.errors = 5
	stats2 := NewStats()
	stats2.errors = 10
	stats.Merge(stats2)
	if stats.errors != 15 {
		t.Errorf("Expected errors to be 15, got %d", stats.errors)
	}
	if stats2.errors != 10 {
		t.Errorf("Expected errors to be 10, got %d", stats2.errors)
	}
}

func TestStatsDuplicatesMerge(t *testing.T) {
	stats := NewStats()
	stats.duplicates = 5
	stats2 := NewStats()
	stats2.duplicates = 10
	stats.Merge(stats2)
	if stats.duplicates != 15 {
		t.Errorf("Expected duplicates to be 15, got %d", stats.duplicates)
	}
	if stats2.duplicates != 10 {
		t.Errorf("Expected duplicates to be 10, got %d", stats2.duplicates)
	}
}
