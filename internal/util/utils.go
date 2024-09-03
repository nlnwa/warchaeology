package util

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/shirou/gopsutil/disk"
	"github.com/spf13/cast"
)

const BadgerRecommendedMaxFileDescr = 65535

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

func DiskFree(path string) (free int64) {
	s, e := disk.Usage(path)
	if e != nil {
		fmt.Printf("ERROR: %v\n", e)
		return 0
	}
	return int64(s.Free)
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
	return OutOfSpaceError(fmt.Sprintf(format, a...))
}

func (o OutOfSpaceError) Error() string {
	return string(o)
}

func StdoutIsTerminal() bool {
	o, _ := os.Stdout.Stat()
	if (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		return true
	} else { //It is not the terminal
		return false
	}
}
