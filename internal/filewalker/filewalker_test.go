package filewalker_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/nlnwa/warchaeology/internal/filewalker"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
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
		{"no suffix", []string{"testdir"}, nil, false, false, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc")}},
		{"one suffix", []string{"testdir"}, []string{".aa"}, false, false, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa")}},
		{"two suffixes", []string{"testdir"}, []string{".aa", ".bb"}, false, false, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb")}},
		{"follow symlinks", []string{"testdir"}, nil, false, true, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc")}},
		{"two dirs", []string{"testdir", "testdir2"}, []string{}, false, false, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc"), filepath.Join("testdir2", "f6.aa")}},
		{"two dirs with symlinks", []string{"testdir", "testdir2"}, nil, false, true, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc"), filepath.Join("testdir2", "f6.aa")}},
		{"file and dir", []string{"testdir", filepath.Join("testdir", "f1.aa")}, []string{}, false, false, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc")}},
		{"file and dir with symlinks", []string{"testdir", filepath.Join("testdir", "f1.aa")}, nil, false, true, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc")}},
		{"recursive no suffix", []string{"testdir"}, nil, true, false, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc"), filepath.Join("testdir", "subdir1/f5.aa")}},
		{"recursive one suffix", []string{"testdir"}, []string{".aa"}, true, false, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "subdir1/f5.aa")}},
		{"recursive two suffixes", []string{"testdir"}, []string{".aa", ".bb"}, true, false, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "subdir1/f5.aa")}},
		{"recursive follow symlinks", []string{"testdir"}, nil, true, true, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc"), filepath.Join("testdir", "subdir1/f5.aa"), filepath.Join("testdir2", "f6.aa")}},
		{"recursive two dirs", []string{"testdir", "testdir2"}, []string{}, true, false, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc"), filepath.Join("testdir", "subdir1/f5.aa"), filepath.Join("testdir2", "f6.aa")}},
		{"recursive two dirs with symlinks", []string{"testdir", "testdir2"}, nil, true, true, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc"), filepath.Join("testdir", "subdir1/f5.aa"), filepath.Join("testdir2", "f6.aa")}},
		{"recursive file and dir", []string{"testdir", filepath.Join("testdir", "f1.aa")}, []string{}, true, false, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc"), filepath.Join("testdir", "subdir1/f5.aa")}},
		{"recursive file and dir with symlinks", []string{"testdir", filepath.Join("testdir", "f1.aa")}, nil, true, true, walker{filepath.Join("testdir", "f1.aa"), filepath.Join("testdir", "f2.aa"), filepath.Join("testdir", "f3.bb"), filepath.Join("testdir", "f4.cc"), filepath.Join("testdir", "subdir1/f5.aa"), filepath.Join("testdir2", "f6.aa")}},
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

func (w *walker) walkfunc(_ afero.Fs, path string) filewalker.Result {
	*w = append(*w, path)
	return filewalker.NewResult(path)
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
