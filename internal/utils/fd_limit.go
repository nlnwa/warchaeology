//go:build !linux

package utils

// CheckFileDescriptorLimit does nothing when OS is not Linux
func CheckFileDescriptorLimit(limit uint64) {
}
