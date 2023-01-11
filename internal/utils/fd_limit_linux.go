//go:build linux

/*
 * Copyright 2023 National Library of Norway.
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
