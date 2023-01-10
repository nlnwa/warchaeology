/*
 * Copyright 2021 National Library of Norway.
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
 */

package internal

import (
	"fmt"
	"github.com/spf13/cast"
	"strings"
	"syscall"
	"unicode"
)

// Contains searches string array s from string e
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// CropString crops a given string if it is bigger than size
func CropString(s string, size int) string {
	if len(s) > size {
		s = s[:size-1] + "\u2026"
	}
	return s
}

// AbsInt64 takes an int64 and returns the absolute value
// Source: http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
func AbsInt64(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

// DiskFree returns the remaining space on disk in bytes
func DiskFree(path string) (free int64) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	free = int64(fs.Bfree) * fs.Bsize
	return
}

const BadgerRecommendedMaxFileDescr = 65535

// CheckFileDescriptorLimit checks if the OS limit for open file descriptors is high enough and if not, tries to increase it.
func CheckFileDescriptorLimit(limit uint64) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Getting Rlimit ", err)
	}

	if rLimit.Max < limit {
		fmt.Printf("It is recomended to set max file descriptors to at least %d. Current value is %d\n", limit, rLimit.Max)
	}

	if rLimit.Cur > limit {
		return
	}

	rLimit.Cur = limit
	if rLimit.Cur > rLimit.Max {
		rLimit.Cur = rLimit.Max
	}

	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Setting Rlimit ", err)
	}
}

// ParseSizeInBytes converts strings like 1GB or 12 mb into an unsigned integer number of bytes
func ParseSizeInBytes(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	lastChar := len(sizeStr) - 1
	multiplier := int64(1)

	if lastChar > 0 {
		if sizeStr[lastChar] == 'b' || sizeStr[lastChar] == 'B' {
			if lastChar > 1 {
				switch unicode.ToLower(rune(sizeStr[lastChar-1])) {
				case 'k':
					multiplier = 1 << 10
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				case 'm':
					multiplier = 1 << 20
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				case 'g':
					multiplier = 1 << 30
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				case 't':
					multiplier = 1 << 40
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				case 'p':
					multiplier = 1 << 50
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				default:
					multiplier = 1
					sizeStr = strings.TrimSpace(sizeStr[:lastChar])
				}
			}
		}
	}

	size := cast.ToInt64(sizeStr)
	if size < 0 {
		size = 0
	}

	return safeMul(size, multiplier)
}

func safeMul(a, b int64) int64 {
	c := a * b
	if a > 1 && b > 1 && c/b != a {
		return 0
	}
	return c
}

type OutOfSpaceError string

func NewOutOfSpaceError(format string, a ...any) OutOfSpaceError {
	var e OutOfSpaceError
	e = OutOfSpaceError(fmt.Sprintf(format, a))
	return e
}

func (o OutOfSpaceError) Error() string {
	return string(o)
}
