package hooks

import (
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	windowsOs = "windows"
)

var testHook = "./testdata/test_hook.sh"

func TestOpenInputFileHook(t *testing.T) {
	if runtime.GOOS == windowsOs {
		t.Skip("TODO (https://github.com/nlnwa/warchaeology/issues/89): This test fails on windows")
	}
	tests := []struct {
		name     string
		command  string
		hook     string
		fileName string
		want     []string
		wantErr  string
	}{
		{"ok", "test", testHook, "test.warc.gz", []string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=OpenInputFile"}, ""},
		{"unknown hook", "test", "test_hook.sh", "test.warc.gz", nil, "executable file 'test_hook.sh' not found in $PATH for OpenInputFileHook"},
		{"exit status error", "test general error", testHook, "test.warc.gz", []string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=OpenInputFile"}, "exit status error"},
		{"exit status skip", "test skip file", testHook, "test.warc.gz", []string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=OpenInputFile"}, ErrSkipFile.Error()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := NewOpenInputFileHook(tt.command, tt.hook)
			if err != nil {
				if tt.wantErr != "" {
					require.Error(t, err)
					assert.Equal(t, tt.wantErr, err.Error())
					return
				} else {
					assert.NoError(t, err)
				}
			}

			out, err := h.Output(tt.fileName)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				return
			} else {
				assert.NoError(t, err)
			}

			env := strings.Split(strings.TrimSpace(string(out)), "\n")
			slices.Sort(env)

			if !slices.Equal(env, tt.want) {
				t.Errorf("OpenInputFileHook.Run() = %v, want %v", env, tt.want)
			}
		})
	}
}

func TestCloseInputFileHook(t *testing.T) {
	if runtime.GOOS == windowsOs {
		t.Skip("TODO (https://github.com/nlnwa/warchaeology/issues/89): This test fails on windows")
	}
	tests := []struct {
		name       string
		command    string
		hook       string
		fileName   string
		errorCount int64
		want       []string
		wantErr    string
	}{
		{"no error", "test", testHook, "test.warc.gz", 0, []string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=CloseInputFile"}, ""},
		{"error", "test", testHook, "test.warc.gz", 2, []string{"WARC_COMMAND=test", "WARC_ERROR_COUNT=2", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=CloseInputFile"}, ""},
		{"unknown hook", "test", "test_hook.sh", "test.warc.gz", 0, nil, "executable file 'test_hook.sh' not found in $PATH for CloseInputFileHook"},
		{"exit status error", "test general error", testHook, "test.warc.gz", 0, []string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=OpenInputFile"}, "exit status error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := NewCloseInputFileHook(tt.command, tt.hook)
			if err != nil {
				if tt.wantErr != "" {
					require.Error(t, err)
					assert.Equal(t, tt.wantErr, err.Error())
					return
				} else {
					assert.NoError(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			out, err := h.Output(tt.fileName, tt.errorCount)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				return
			} else {
				assert.NoError(t, err)
			}

			env := strings.Split(strings.TrimSpace(string(out)), "\n")
			slices.Sort(env)

			if !slices.Equal(env, tt.want) {
				t.Errorf("CloseInputFileHook.Run() = %v, want %v", env, tt.want)
			}
		})
	}
}

func TestOpenOutputFileHook(t *testing.T) {
	if runtime.GOOS == windowsOs {
		t.Skip("TODO (https://github.com/nlnwa/warchaeology/issues/89): This test fails on windows")
	}
	tests := []struct {
		name        string
		command     string
		hook        string
		fileName    string
		srcFileName string
		want        []string
		wantErr     string
	}{
		{"no srcFile", "test", testHook, "test.warc.gz", "", []string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=OpenOutputFile"}, ""},
		{"srcFile", "test", testHook, "test.warc.gz", "TestSource", []string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=OpenOutputFile", "WARC_SRC_FILE_NAME=TestSource"}, ""},
		{"unknown hook", "test", "test_hook.sh", "test.warc.gz", "", nil, "executable file 'test_hook.sh' not found in $PATH for OpenOutputFileHook"},
		{"exit status error", "test general error", testHook, "test.warc.gz", "", []string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=OpenInputFile"}, "exit status error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := NewOpenOutputFileHook(tt.command, tt.hook)
			if err != nil {
				if tt.wantErr != "" {
					require.Error(t, err)
					assert.Equal(t, tt.wantErr, err.Error())
					return
				} else {
					assert.NoError(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			h = h.WithSrcFileName(tt.srcFileName)
			out, err := h.Output(tt.fileName)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				return
			} else {
				assert.NoError(t, err)
			}

			env := strings.Split(strings.TrimSpace(string(out)), "\n")
			slices.Sort(env)

			if !slices.Equal(env, tt.want) {
				t.Errorf("OpenOutputFileHook.Run() = %v, want %v", env, tt.want)
			}
		})
	}
}

func TestCloseOutputFileHook(t *testing.T) {
	if runtime.GOOS == windowsOs {
		t.Skip("TODO (https://github.com/nlnwa/warchaeology/issues/89): This test fails on windows")
	}
	tests := []struct {
		name        string
		command     string
		hook        string
		fileName    string
		size        int64
		warcInfoId  string
		srcFileName string
		errorCount  int64
		want        []string
		wantErr     string
	}{
		{"no extras", "test", testHook, "test.warc.gz",
			1234, "", "", 0,
			[]string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=CloseOutputFile", "WARC_SIZE=1234"},
			""},
		{"error", "test", testHook, "test.warc.gz",
			1234, "", "", 2,
			[]string{"WARC_COMMAND=test", "WARC_ERROR_COUNT=2", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=CloseOutputFile", "WARC_SIZE=1234"},
			""},
		{"warcInfoId", "test", testHook, "test.warc.gz",
			1234, "TestId", "", 0,
			[]string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=CloseOutputFile", "WARC_INFO_ID=TestId", "WARC_SIZE=1234"},
			""},
		{"srcFile", "test", testHook, "test.warc.gz",
			1234, "", "TestSource", 0,
			[]string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=CloseOutputFile", "WARC_SIZE=1234", "WARC_SRC_FILE_NAME=TestSource"},
			""},
		{"warcInfoId + srcFile", "test", testHook, "test.warc.gz",
			1234, "TestId", "TestSource", 0,
			[]string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=CloseOutputFile", "WARC_INFO_ID=TestId", "WARC_SIZE=1234", "WARC_SRC_FILE_NAME=TestSource"},
			""},
		{"unknown hook", "test", "test_hook.sh", "test.warc.gz",
			0, "", "", 0, nil, "executable file 'test_hook.sh' not found in $PATH for CloseOutputFileHook"},
		{"exit status error", "test general error", testHook, "test.warc.gz",
			0, "", "", 0, []string{"WARC_COMMAND=test", "WARC_FILE_NAME=test.warc.gz", "WARC_HOOK_TYPE=OpenInputFile"}, "exit status error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := NewCloseOutputFileHook(tt.command, tt.hook)
			if err != nil {
				if tt.wantErr != "" {
					require.Error(t, err)
					assert.Equal(t, tt.wantErr, err.Error())
					return
				} else {
					assert.NoError(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			h = h.WithSrcFileName(tt.srcFileName).WithErrorCount(tt.errorCount)
			out, err := h.Output(tt.fileName, tt.size, tt.warcInfoId)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				return
			} else {
				assert.NoError(t, err)
			}

			env := strings.Split(strings.TrimSpace(string(out)), "\n")
			slices.Sort(env)

			if !slices.Equal(env, tt.want) {
				t.Errorf("CloseOutputFileHook.Run() = %v, want %v", env, tt.want)
			}
		})
	}
}

func Test_checkExists(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		wantExistence bool
	}{
		{"test_hook.sh", "test_hook.sh", false},
		{testHook, testHook, true},
		{"ls", "ls", true},
		{"foo", "foo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if checkExists(tt.command) != tt.wantExistence {
				t.Errorf("checkExists() = %v, want %v", checkExists(tt.command), tt.wantExistence)
			}
		})
	}
}
