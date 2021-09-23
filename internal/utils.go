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

// Searches string array s from string e
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Crops a given string if it is bigger than size
func CropString(s string, size int) string {
	if len(s) > size {
		s = s[:size-3] + "..."
	}
	return s
}

// Source: http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html:
//
// Takes an int64 and returns the absolute value
func AbsInt64(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}
