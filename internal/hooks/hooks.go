package hooks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

const (
	EnvCommand     = "WARC_COMMAND"
	EnvFileName    = "WARC_FILE_NAME"
	EnvErrorCount  = "WARC_ERROR_COUNT"
	EnvWarcInfoId  = "WARC_INFO_ID"
	EnvFileSize    = "WARC_SIZE"
	EnvHash        = "WARC_HASH"
	EnvSrcFileName = "WARC_SRC_FILE_NAME"
	EnvHookType    = "WARC_HOOK_TYPE"
)

var ErrSkipFile = fmt.Errorf("skip file")

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
		return OpenInputFileHook{}, ErrCommandNotFound{"OpenInputFile", hook}
	}

	openInputFileHook := OpenInputFileHook{
		hook: hook,
		cmd:  cmd,
	}
	return openInputFileHook, nil
}

func (openInputFileHook OpenInputFileHook) Run(fileName string) error {
	if openInputFileHook.hook == "" {
		return nil
	}
	outputBytes, err := openInputFileHook.Output(fileName)
	if len(outputBytes) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(outputBytes))
	}
	return err
}

func (openInputFileHook OpenInputFileHook) Output(fileName string) ([]byte, error) {
	if openInputFileHook.hook == "" {
		return nil, nil
	}

	cmd := exec.Command(openInputFileHook.hook)
	cmd.Env = append(cmd.Environ(), EnvCommand+"="+openInputFileHook.cmd)
	cmd.Env = append(cmd.Environ(), EnvHookType+"=OpenInputFile")
	cmd.Env = append(cmd.Environ(), EnvFileName+"="+fileName)

	outputBytes, err := cmd.CombinedOutput()
	outputBytes = bytes.TrimSpace(outputBytes)
	if exitError, ok := err.(*exec.ExitError); ok {
		switch exitError.ExitCode() {
		case 1:
			return nil, fmt.Errorf("%s", outputBytes)
		case 10:
			return nil, ErrSkipFile
		}
	}
	return outputBytes, err
}

type CloseInputFileHook struct {
	cmd  string
	hook string
	hash string
}

// NewCloseInputFileHook creates a new CloseInputFileHook
//
// cmd is the current warc subcommand.
// hook is the command/script to execute.
// returns a CloseInputFileHook or ErrCommandNotFound if the hook command is not found
func NewCloseInputFileHook(cmd, hook string) (CloseInputFileHook, error) {
	if !checkExists(hook) {
		return CloseInputFileHook{}, ErrCommandNotFound{"CloseInputFile", hook}
	}

	closeInputFileHook := CloseInputFileHook{
		cmd:  cmd,
		hook: hook,
	}
	return closeInputFileHook, nil
}

func (closeInputFileHook CloseInputFileHook) WithHash(hash string) CloseInputFileHook {
	closeInputFileHook.hash = hash
	return closeInputFileHook
}

func (closeInputFileHook CloseInputFileHook) Run(fileName string, errorCount int64) error {
	if closeInputFileHook.hook == "" {
		return nil
	}
	outputBytes, err := closeInputFileHook.Output(fileName, errorCount)
	if len(outputBytes) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(outputBytes))
	}
	return err
}

func (closeInputFileHook CloseInputFileHook) Output(fileName string, errorCount int64) ([]byte, error) {
	if closeInputFileHook.hook == "" {
		return nil, nil
	}

	cmd := exec.Command(closeInputFileHook.hook)
	cmd.Env = append(cmd.Environ(), EnvCommand+"="+closeInputFileHook.cmd)
	cmd.Env = append(cmd.Environ(), EnvHookType+"=CloseInputFile")
	cmd.Env = append(cmd.Environ(), EnvFileName+"="+fileName)
	if errorCount > 0 {
		cmd.Env = append(cmd.Environ(), fmt.Sprintf("%s=%d", EnvErrorCount, errorCount))
	}
	if closeInputFileHook.hash != "" {
		cmd.Env = append(cmd.Environ(), EnvHash+"="+closeInputFileHook.hash)
	}

	outputBytes, err := cmd.CombinedOutput()
	outputBytes = bytes.TrimSpace(outputBytes)
	if exitError, ok := err.(*exec.ExitError); ok {
		switch exitError.ExitCode() {
		case 1:
			return nil, fmt.Errorf("%s", outputBytes)
		}
	}
	return outputBytes, err
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
		return OpenOutputFileHook{}, ErrCommandNotFound{"OpenOutputFile", hook}
	}

	openOutputFileHook := OpenOutputFileHook{
		cmd:  cmd,
		hook: hook,
	}
	return openOutputFileHook, nil
}

func (openOutputFileHook OpenOutputFileHook) WithSrcFileName(srcFileName string) OpenOutputFileHook {
	openOutputFileHook.srcFileName = srcFileName
	return openOutputFileHook
}

func (openOutputFileHook OpenOutputFileHook) Run(fileName string) error {
	if openOutputFileHook.hook == "" {
		return nil
	}
	outputBytes, err := openOutputFileHook.Output(fileName)
	if len(outputBytes) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(outputBytes))
	}
	return err
}

func (openOutputFileHook OpenOutputFileHook) Output(fileName string) ([]byte, error) {
	if openOutputFileHook.hook == "" {
		return nil, nil
	}

	cmd := exec.Command(openOutputFileHook.hook)
	cmd.Env = append(cmd.Environ(), EnvCommand+"="+openOutputFileHook.cmd)
	cmd.Env = append(cmd.Environ(), EnvHookType+"=OpenOutputFile")
	cmd.Env = append(cmd.Environ(), EnvFileName+"="+fileName)
	if openOutputFileHook.srcFileName != "" {
		cmd.Env = append(cmd.Environ(), EnvSrcFileName+"="+openOutputFileHook.srcFileName)
	}

	outputBytes, err := cmd.CombinedOutput()
	outputBytes = bytes.TrimSpace(outputBytes)
	if exitError, ok := err.(*exec.ExitError); ok {
		switch exitError.ExitCode() {
		case 1:
			return nil, fmt.Errorf("%s", outputBytes)
		}
	}
	return outputBytes, err
}

type CloseOutputFileHook struct {
	cmd         string
	hook        string
	srcFileName string
	hash        string
	errorCount  int64
}

// NewCloseOutputFileHook creates a new CloseOutputFileHook
//
// cmd is the current warc subcommand.
// hook is the command/script to execute.
// returns a CloseOutputFileHook or ErrCommandNotFound if the hook command is not found
func NewCloseOutputFileHook(cmd, hook string) (CloseOutputFileHook, error) {
	if !checkExists(hook) {
		return CloseOutputFileHook{}, ErrCommandNotFound{"CloseOutputFile", hook}
	}

	closeOutputFileHook := CloseOutputFileHook{
		cmd:  cmd,
		hook: hook,
	}
	return closeOutputFileHook, nil
}

func (closeOutputFileHook CloseOutputFileHook) WithSrcFileName(srcFileName string) CloseOutputFileHook {
	closeOutputFileHook.srcFileName = srcFileName
	return closeOutputFileHook
}

func (closeOutputFileHook CloseOutputFileHook) WithHash(hash string) CloseOutputFileHook {
	closeOutputFileHook.hash = hash
	return closeOutputFileHook
}

func (closeOutputFileHook CloseOutputFileHook) WithErrorCount(errorCount int64) CloseOutputFileHook {
	closeOutputFileHook.errorCount = errorCount
	return closeOutputFileHook
}

func (closeOutputFileHook CloseOutputFileHook) Run(fileName string, size int64, warcInfoId string) error {
	if closeOutputFileHook.hook == "" {
		return nil
	}
	outputBytes, err := closeOutputFileHook.Output(fileName, size, warcInfoId)
	if len(outputBytes) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(outputBytes))
	}
	return err
}

func (closeOutputFileHook CloseOutputFileHook) Output(fileName string, size int64, warcInfoId string) ([]byte, error) {
	if closeOutputFileHook.hook == "" {
		return nil, nil
	}

	cmd := exec.Command(closeOutputFileHook.hook)
	cmd.Env = append(cmd.Environ(), EnvCommand+"="+closeOutputFileHook.cmd)
	cmd.Env = append(cmd.Environ(), EnvHookType+"=CloseOutputFile")
	cmd.Env = append(cmd.Environ(), EnvFileName+"="+fileName)
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("%s=%d", EnvFileSize, size))
	if warcInfoId != "" {
		cmd.Env = append(cmd.Environ(), EnvWarcInfoId+"="+warcInfoId)
	}
	if closeOutputFileHook.srcFileName != "" {
		cmd.Env = append(cmd.Environ(), EnvSrcFileName+"="+closeOutputFileHook.srcFileName)
	}
	if closeOutputFileHook.hash != "" {
		cmd.Env = append(cmd.Environ(), EnvHash+"="+closeOutputFileHook.hash)
	}
	if closeOutputFileHook.errorCount > 0 {
		cmd.Env = append(cmd.Environ(), fmt.Sprintf("%s=%d", EnvErrorCount, closeOutputFileHook.errorCount))
	}

	outputBytes, err := cmd.CombinedOutput()
	outputBytes = bytes.TrimSpace(outputBytes)
	if exitError, ok := err.(*exec.ExitError); ok {
		switch exitError.ExitCode() {
		case 1:
			return nil, fmt.Errorf("%s", outputBytes)
		}
	}
	return outputBytes, err
}

func checkExists(command string) bool {
	if command == "" {
		return true
	}
	_, err := exec.LookPath(command)
	return err == nil
}

type ErrCommandNotFound struct {
	hookType string
	command  string
}

func (errCommandNotFound ErrCommandNotFound) Error() string {
	return fmt.Sprintf("executable file '%s' not found in $PATH for %sHook", errCommandNotFound.command, errCommandNotFound.hookType)
}
