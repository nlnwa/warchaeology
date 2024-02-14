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

package filewalker

// StringSet is a set of strings
type StringSet map[string]struct{}

// NewStringSet creates a new StringSet
func NewStringSet() StringSet {
	return make(StringSet)
}

// Add adds a string to the set
func (s StringSet) Add(path string) {
	s[path] = struct{}{}
}

// Contains checks if a string is in the set
func (s StringSet) Contains(path string) bool {
	_, ok := s[path]
	return ok
}
