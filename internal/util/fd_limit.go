//go:build !linux

package util

// CheckFileDescriptorLimit does nothing when OS is not Linux
func CheckFileDescriptorLimit(limit uint64) error {
	return nil
}
