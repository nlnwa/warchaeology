/*
 * Copyright 2024 National Library of Norway.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package hooks

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

var ErrCommandNotFound = errors.New("command not found")

const (
	EnvCommand     = "WARC_COMMAND"
	EnvFileName    = "WARC_FILE_NAME"
	EnvError       = "WARC_ERROR_COUNT"
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
		return OpenInputFileHook{}, ErrCommandNotFound
	}

	h := OpenInputFileHook{
		hook: hook,
		cmd:  cmd,
	}
	return h, nil
}

func (h OpenInputFileHook) Run(fileName string) error {
	b, err := h.Output(fileName)
	if b = bytes.TrimSpace(b); len(b) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(b))
	}
	return err
}

func (h OpenInputFileHook) Output(fileName string) ([]byte, error) {
	if h.hook == "" {
		return nil, nil
	}

	c := exec.Command(h.hook)
	c.Env = append(c.Environ(), EnvCommand+"="+h.cmd)
	c.Env = append(c.Environ(), EnvFileName+"="+fileName)
	c.Env = append(c.Environ(), EnvHookType+"=OpenInputFile")

	return c.CombinedOutput()
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
		return CloseInputFileHook{}, ErrCommandNotFound
	}

	h := CloseInputFileHook{
		cmd:  cmd,
		hook: hook,
	}
	return h, nil
}

func (h CloseInputFileHook) WithHash(hash string) CloseInputFileHook {
	h.hash = hash
	return h
}

func (h CloseInputFileHook) Run(fileName string, errorCount int64) error {
	b, err := h.Output(fileName, errorCount)
	if b = bytes.TrimSpace(b); len(b) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(b))
	}
	return err
}

func (h CloseInputFileHook) Output(fileName string, errorCount int64) ([]byte, error) {
	if h.hook == "" {
		return nil, nil
	}

	c := exec.Command(h.hook)
	c.Env = append(c.Environ(), EnvCommand+"="+h.cmd)
	c.Env = append(c.Environ(), EnvFileName+"="+fileName)
	if errorCount > 0 {
		c.Env = append(c.Environ(), fmt.Sprintf("%s=%d", EnvError, errorCount))
	}
	if h.hash != "" {
		c.Env = append(c.Environ(), EnvHash+"="+h.hash)
	}
	c.Env = append(c.Environ(), EnvHookType+"=CloseInputFile")

	return c.CombinedOutput()
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
		return OpenOutputFileHook{}, ErrCommandNotFound
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
	b, err := h.Output(fileName)
	if b = bytes.TrimSpace(b); len(b) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(b))
	}
	return err
}

func (h OpenOutputFileHook) Output(fileName string) ([]byte, error) {
	if h.hook == "" {
		return nil, nil
	}

	c := exec.Command(h.hook)
	c.Env = append(c.Environ(), EnvCommand+"="+h.cmd)
	c.Env = append(c.Environ(), EnvFileName+"="+fileName)
	if h.srcFileName != "" {
		c.Env = append(c.Environ(), EnvSrcFileName+"="+h.srcFileName)
	}
	c.Env = append(c.Environ(), EnvHookType+"=OpenOutputFile")

	return c.CombinedOutput()
}

type CloseOutputFileHook struct {
	cmd         string
	hook        string
	srcFileName string
	hash        string
}

// NewCloseOutputFileHook creates a new CloseOutputFileHook
//
// cmd is the current warc subcommand.
// hook is the command/script to execute.
// returns a CloseOutputFileHook or ErrCommandNotFound if the hook command is not found
func NewCloseOutputFileHook(cmd, hook string) (CloseOutputFileHook, error) {
	if !checkExists(hook) {
		return CloseOutputFileHook{}, ErrCommandNotFound
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

func (h CloseOutputFileHook) WithHash(hash string) CloseOutputFileHook {
	h.hash = hash
	return h
}

func (h CloseOutputFileHook) Run(fileName string, size int64, warcInfoId string) error {
	b, err := h.Output(fileName, size, warcInfoId)
	if b = bytes.TrimSpace(b); len(b) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, string(b))
	}
	return err
}

func (h CloseOutputFileHook) Output(fileName string, size int64, warcInfoId string) ([]byte, error) {
	if h.hook == "" {
		return nil, nil
	}

	c := exec.Command(h.hook)
	c.Env = append(c.Environ(), EnvCommand+"="+h.cmd)
	c.Env = append(c.Environ(), EnvFileName+"="+fileName)
	c.Env = append(c.Environ(), fmt.Sprintf("%s=%d", EnvFileSize, size))
	if warcInfoId != "" {
		c.Env = append(c.Environ(), EnvWarcInfoId+"="+warcInfoId)
	}
	if h.srcFileName != "" {
		c.Env = append(c.Environ(), EnvSrcFileName+"="+h.srcFileName)
	}
	if h.hash != "" {
		c.Env = append(c.Environ(), EnvHash+"="+h.hash)
	}
	c.Env = append(c.Environ(), EnvHookType+"=CloseOutputFile")

	return c.CombinedOutput()
}

func checkExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
