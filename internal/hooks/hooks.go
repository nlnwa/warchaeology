package hooks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/nlnwa/warchaeology/v4/internal/stat"
)

const (
	EnvCommand     = "WARC_COMMAND"
	EnvFileName    = "WARC_FILE_NAME"
	EnvErrorCount  = "WARC_ERROR_COUNT"
	EnvError       = "WARC_ERROR"
	EnvWarcInfoId  = "WARC_INFO_ID"
	EnvFileSize    = "WARC_SIZE"
	EnvHash        = "WARC_HASH"
	EnvSrcFileName = "WARC_SRC_FILE_NAME"
	EnvHookType    = "WARC_HOOK_TYPE"
)

type OpenInputFileHook struct {
	cmd  string
	hook string
}

// NewOpenInputFileHook creates a new OpenInputFileHook
//
// cmd is the current warc subcommand.
// hook is the command/script to execute.
// returns an OpenInputFileHook or ErrCommandNotFound if the hook command is not found
func NewOpenInputFileHook(cmd, hook string) (OpenInputFileHook, error) {
	if !checkExists(hook) {
		return OpenInputFileHook{}, CommandNotFoundError{"OpenInputFile", hook}
	}

	return OpenInputFileHook{
		hook: hook,
		cmd:  cmd,
	}, nil
}

func (h OpenInputFileHook) Run(fileName string) error {
	if h.hook == "" {
		return nil
	}
	b, err := h.Output(fileName)
	if len(b) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(b))
	}
	return err
}

func (h OpenInputFileHook) Output(fileName string) ([]byte, error) {
	c := exec.CommandContext(context.TODO(), h.hook)
	c.Env = append(c.Environ(), EnvCommand+"="+h.cmd)
	c.Env = append(c.Environ(), EnvHookType+"=OpenInputFile")
	c.Env = append(c.Environ(), EnvFileName+"="+fileName)

	b, err := c.CombinedOutput()
	b = bytes.TrimSpace(b)
	if e, ok := err.(*exec.ExitError); ok {
		switch e.ExitCode() {
		case 1:
			return nil, fmt.Errorf("%s", b)
		case 10:
			return nil, ErrSkipFile
		}
	}
	return b, err
}

type CloseInputFileHook struct {
	cmd  string
	hook string
}

// NewCloseInputFileHook creates a new CloseInputFileHook
//
// cmd is the current warc subcommand.
// hook is the command/script to execute.
// returns a CloseInputFileHook or ErrCommandNotFound if the hook command is not found
func NewCloseInputFileHook(cmd, hook string) (CloseInputFileHook, error) {
	if !checkExists(hook) {
		return CloseInputFileHook{}, CommandNotFoundError{"CloseInputFile", hook}
	}

	h := CloseInputFileHook{
		cmd:  cmd,
		hook: hook,
	}
	return h, nil
}

func (h CloseInputFileHook) Run(fileName string, result stat.Result, resultErr error) error {
	if h.hook == "" {
		return nil
	}
	b, err := h.Output(fileName, result, resultErr)
	if len(b) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(b))
	}
	return err
}

func (h CloseInputFileHook) Output(fileName string, result stat.Result, resultErr error) ([]byte, error) {
	c := exec.CommandContext(context.TODO(), h.hook)
	c.Env = append(c.Environ(), EnvCommand+"="+h.cmd)
	c.Env = append(c.Environ(), EnvHookType+"=CloseInputFile")
	c.Env = append(c.Environ(), EnvFileName+"="+fileName)
	if resultErr != nil {
		c.Env = append(c.Environ(), EnvError+"="+resultErr.Error())
	}
	if result != nil {
		errorCount := result.ErrorCount()
		if errorCount > 0 {
			c.Env = append(c.Environ(), fmt.Sprintf("%s=%d", EnvErrorCount, errorCount))
		}

		hash := result.Hash()
		if hash != "" {
			c.Env = append(c.Environ(), EnvHash+"="+hash)
		}
	}

	b, err := c.CombinedOutput()
	b = bytes.TrimSpace(b)
	if e, ok := err.(*exec.ExitError); ok {
		switch e.ExitCode() {
		case 1:
			return nil, fmt.Errorf("%s", b)
		}
	}
	return b, err
}

type OpenOutputFileHook struct {
	cmd         string
	hook        string
	srcFileName string
}

// NewOpenOutputFileHook creates a new OpenOutputFileHook
//
// cmd is the current warc subcommand.
// hook is the command/script to execute.
// returns an OpenOutputFileHook or ErrCommandNotFound if the hook command is not found
func NewOpenOutputFileHook(cmd, hook string) (OpenOutputFileHook, error) {
	if !checkExists(hook) {
		return OpenOutputFileHook{}, CommandNotFoundError{"OpenOutputFile", hook}
	}

	h := OpenOutputFileHook{
		cmd:  cmd,
		hook: hook,
	}
	return h, nil
}

func (h OpenOutputFileHook) WithSrcFileName(srcFileName string) OpenOutputFileHook {
	h.srcFileName = srcFileName
	return h
}

func (h OpenOutputFileHook) Run(fileName string) error {
	if h.hook == "" {
		return nil
	}
	b, err := h.Output(fileName)
	if len(b) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(b))
	}
	return err
}

func (h OpenOutputFileHook) Output(fileName string) ([]byte, error) {
	c := exec.CommandContext(context.TODO(), h.hook)
	c.Env = append(c.Environ(), EnvCommand+"="+h.cmd)
	c.Env = append(c.Environ(), EnvHookType+"=OpenOutputFile")
	c.Env = append(c.Environ(), EnvFileName+"="+fileName)
	if h.srcFileName != "" {
		c.Env = append(c.Environ(), EnvSrcFileName+"="+h.srcFileName)
	}

	b, err := c.CombinedOutput()
	b = bytes.TrimSpace(b)
	if e, ok := err.(*exec.ExitError); ok {
		switch e.ExitCode() {
		case 1:
			return nil, fmt.Errorf("%s", b)
		}
	}
	return b, err
}

type CloseOutputFileHook struct {
	cmd         string
	hook        string
	srcFileName string
}

// NewCloseOutputFileHook creates a new CloseOutputFileHook
//
// cmd is the current warc subcommand.
// hook is the command/script to execute.
// returns a CloseOutputFileHook or ErrCommandNotFound if the hook command is not found
func NewCloseOutputFileHook(cmd, hook string) (CloseOutputFileHook, error) {
	if !checkExists(hook) {
		return CloseOutputFileHook{}, CommandNotFoundError{"CloseOutputFile", hook}
	}

	h := CloseOutputFileHook{
		cmd:  cmd,
		hook: hook,
	}
	return h, nil
}

func (h CloseOutputFileHook) WithSrcFileName(srcFileName string) CloseOutputFileHook {
	h.srcFileName = srcFileName
	return h
}

func (h CloseOutputFileHook) Run(fileName string, size int64, warcInfoId string) error {
	if h.hook == "" {
		return nil
	}
	b, err := h.Output(fileName, size, warcInfoId)
	if len(b) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(b))
	}
	return err
}

func (h CloseOutputFileHook) Output(fileName string, size int64, warcInfoId string) ([]byte, error) {
	if h.hook == "" {
		return nil, nil
	}

	c := exec.CommandContext(context.TODO(), h.hook)
	c.Env = append(c.Environ(), EnvCommand+"="+h.cmd)
	c.Env = append(c.Environ(), EnvHookType+"=CloseOutputFile")
	c.Env = append(c.Environ(), EnvFileName+"="+fileName)
	c.Env = append(c.Environ(), fmt.Sprintf("%s=%d", EnvFileSize, size))
	if warcInfoId != "" {
		c.Env = append(c.Environ(), EnvWarcInfoId+"="+warcInfoId)
	}
	if h.srcFileName != "" {
		c.Env = append(c.Environ(), EnvSrcFileName+"="+h.srcFileName)
	}

	b, err := c.CombinedOutput()
	b = bytes.TrimSpace(b)
	if e, ok := err.(*exec.ExitError); ok {
		switch e.ExitCode() {
		case 1:
			return nil, fmt.Errorf("%s", b)
		}
	}
	return b, err
}

func checkExists(command string) bool {
	if command == "" {
		return true
	}
	_, err := exec.LookPath(command)
	return err == nil
}

type CommandNotFoundError struct {
	hookType string
	command  string
}

func (e CommandNotFoundError) Error() string {
	return fmt.Sprintf("executable file '%s' not found in $PATH for %sHook", e.command, e.hookType)
}

var ErrSkipFile = errors.New("skip file")
