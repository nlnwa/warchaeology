package filewalker_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/nlnwa/warchaeology/v3/internal/filewalker"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var (
	testDir  = filepath.Join("testdata", "testdir")
	testDir2 = filepath.Join("testdata", "testdir2")
)

func TestFilewalker_Walk(t *testing.T) {
	tests := []struct {
		name           string
		paths          []string
		suffixes       []string
		recursive      bool
		followSymlinks bool
		expected       []string
	}{
		{"non existing dir", []string{"aa"}, nil, true, false, nil},
		{"empty dir", []string{""}, nil, true, false, nil},
		{"no suffix", []string{testDir}, nil, false, false, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc")}},
		{"one suffix", []string{testDir}, []string{".aa"}, false, false, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa")}},
		{"two suffixes", []string{testDir}, []string{".aa", ".bb"}, false, false, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb")}},
		{"follow symlinks", []string{testDir}, nil, false, true, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc")}},
		{"two dirs", []string{testDir, testDir2}, []string{}, false, false, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc"), filepath.Join(testDir2, "f6.aa")}},
		{"two dirs with symlinks", []string{testDir, testDir2}, nil, false, true, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc"), filepath.Join(testDir2, "f6.aa")}},
		{"file and dir", []string{testDir, filepath.Join(testDir, "f1.aa")}, []string{}, false, false, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc")}},
		{"file and dir with symlinks", []string{testDir, filepath.Join(testDir, "f1.aa")}, nil, false, true, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc")}},
		{"recursive no suffix", []string{testDir}, nil, true, false, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc"), filepath.Join(testDir, "subdir1/f5.aa")}},
		{"recursive one suffix", []string{testDir}, []string{".aa"}, true, false, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "subdir1/f5.aa")}},
		{"recursive two suffixes", []string{testDir}, []string{".aa", ".bb"}, true, false, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "subdir1/f5.aa")}},
		{"recursive follow symlinks", []string{testDir}, nil, true, true, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc"), filepath.Join(testDir, "subdir1/f5.aa"), filepath.Join(testDir2, "f6.aa")}},
		{"recursive two dirs", []string{testDir, testDir2}, []string{}, true, false, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc"), filepath.Join(testDir, "subdir1/f5.aa"), filepath.Join(testDir2, "f6.aa")}},
		{"recursive two dirs with symlinks", []string{testDir, testDir2}, nil, true, true, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc"), filepath.Join(testDir, "subdir1/f5.aa"), filepath.Join(testDir2, "f6.aa")}},
		{"recursive file and dir", []string{testDir, filepath.Join(testDir, "f1.aa")}, []string{}, true, false, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc"), filepath.Join(testDir, "subdir1/f5.aa")}},
		{"recursive file and dir with symlinks", []string{testDir, filepath.Join(testDir, "f1.aa")}, nil, true, true, []string{filepath.Join(testDir, "f1.aa"), filepath.Join(testDir, "f2.aa"), filepath.Join(testDir, "f3.bb"), filepath.Join(testDir, "f4.cc"), filepath.Join(testDir, "subdir1/f5.aa"), filepath.Join(testDir2, "f6.aa")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := filewalker.New(
				filewalker.WithFs(afero.NewOsFs()),
				filewalker.WithFollowSymlinks(tt.followSymlinks),
				filewalker.WithRecursive(tt.recursive),
				filewalker.WithSuffixes(tt.suffixes),
			)

			var got []string
			walkfunc := func(_ afero.Fs, path string, err error) error {
				if err != nil {
					return nil
				}
				got = append(got, path)
				return nil
			}

			for _, path := range tt.paths {
				err := f.Walk(context.Background(), path, walkfunc)
				if err != nil {
					t.Error(err)
				}
			}

			assert.Equal(t, tt.expected, got)
		})
	}
}
