package filewalker

import (
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
			f := New(tt.paths, tt.recursive, tt.followSymlinks, tt.suffixes, w.walkfunc)
			err := f.Walk()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, w)
		})
	}
}

type walker []string

func (w *walker) walkfunc(path string) {
	*w = append(*w, path)
}
