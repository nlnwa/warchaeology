package filewalker_test

import (
	"context"
	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilewalker_Walk(t *testing.T) {
	tests := []struct {
		name           string
		paths          []string
		suffixes       []string
		recursive      bool
		followSymlinks bool
		expected       walker
	}{
		{"non existing dir", []string{"aa"}, nil, true, false, walker{}},
		{"empty dir", []string{""}, nil, true, false, walker{}},
		{"no suffix", []string{"testdir"}, nil, false, false, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc"}},
		{"one suffix", []string{"testdir"}, []string{".aa"}, false, false, walker{"testdir/f1.aa", "testdir/f2.aa"}},
		{"two suffixes", []string{"testdir"}, []string{".aa", ".bb"}, false, false, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb"}},
		{"follow symlinks", []string{"testdir"}, nil, false, true, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc"}},
		{"two dirs", []string{"testdir", "testdir2"}, []string{}, false, false, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc", "testdir2/f6.aa"}},
		{"two dirs with symlinks", []string{"testdir", "testdir2"}, nil, false, true, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc", "testdir2/f6.aa"}},
		{"file and dir", []string{"testdir", "testdir/f1.aa"}, []string{}, false, false, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc", "testdir/f1.aa"}},
		{"file and dir with symlinks", []string{"testdir", "testdir/f1.aa"}, nil, false, true, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc", "testdir/f1.aa"}},
		{"recursive no suffix", []string{"testdir"}, nil, true, false, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc", "testdir/subdir1/f5.aa"}},
		{"recursive one suffix", []string{"testdir"}, []string{".aa"}, true, false, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/subdir1/f5.aa"}},
		{"recursive two suffixes", []string{"testdir"}, []string{".aa", ".bb"}, true, false, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/subdir1/f5.aa"}},
		{"recursive follow symlinks", []string{"testdir"}, nil, true, true, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc", "testdir/subdir1/f5.aa", "testdir2/f6.aa"}},
		{"recursive two dirs", []string{"testdir", "testdir2"}, []string{}, true, false, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc", "testdir/subdir1/f5.aa", "testdir2/f6.aa"}},
		{"recursive two dirs with symlinks", []string{"testdir", "testdir2"}, nil, true, true, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc", "testdir/subdir1/f5.aa", "testdir2/f6.aa"}},
		{"recursive file and dir", []string{"testdir", "testdir/f1.aa"}, []string{}, true, false, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc", "testdir/subdir1/f5.aa", "testdir/f1.aa"}},
		{"recursive file and dir with symlinks", []string{"testdir", "testdir/f1.aa"}, nil, true, true, walker{"testdir/f1.aa", "testdir/f2.aa", "testdir/f3.bb", "testdir/f4.cc", "testdir/subdir1/f5.aa", "testdir2/f6.aa", "testdir/f1.aa"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := walker{}
			f := filewalker.New(tt.paths, tt.recursive, tt.followSymlinks, tt.suffixes, 1, w.walkfunc)
			ctx := context.TODO()
			stats := &result{}
			err := f.Walk(ctx, stats)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, w)
		})
	}
}

type walker []string

func (w *walker) walkfunc(path string) filewalker.Result {
	*w = append(*w, path)
	return &result{}
}

type result struct{}

func (r *result) IncrRecords() {
}

func (r *result) IncrProcessed() {
}

func (r *result) AddError(err error) {
}

func (r *result) Records() int64 {
	return 0
}

func (r *result) Processed() int64 {
	return 0
}

func (r *result) ErrorCount() int64 {
	return 0
}

func (r *result) Errors() []error {
	return []error{}
}

func (r *result) Error() string {
	return ""
}

func (r *result) Merge(s filewalker.Stats) {
}

func (r *result) String() string {
	return ""
}

func (r *result) Log(fileNum int) string {
	return ""
}

func (r *result) GetStats() filewalker.Stats {
	return r
}
