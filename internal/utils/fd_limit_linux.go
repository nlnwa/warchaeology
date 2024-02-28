//go:build linux

package utils

import (
	"fmt"
	"syscall"
)

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
