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

package utils_test

import (
	"github.com/nlnwa/warchaeology/internal/utils"
	"testing"
)

func TestContainsMatchOnMatching(t *testing.T) {
	needle := "needle"
	haystack := []string{"hay1", "hay2", needle, "hay3", "hay5"}

	if !utils.Contains(haystack, needle) {
		t.Error("Failed to find needle")
	}
}

func TestContainsNonMatchOnNonMatching(t *testing.T) {
	needle := "needle"
	haystack := []string{"hay1", "hay2", "hay4", "hay3", "hay5"}

	if utils.Contains(haystack, needle) {
		t.Error("Found element that does not exist in slice")
	}
}

func TestCropStringCrops(t *testing.T) {
	input := "123456789"
	expected := "1234\u2026"

	actual := utils.CropString(input, 5)

	if actual != expected {
		t.Errorf("Expected %s got %s", expected, actual)
	}
}

func TestAbsInt64NegativeInput(t *testing.T) {
	const input int64 = -10
	actual := utils.AbsInt64(input)

	const expected int64 = 10
	if actual != expected {
		t.Errorf("Unexpected value. expected %d got %d", expected, actual)
	}
}

func TestAbsInt64PositiveInput(t *testing.T) {
	const input int64 = 10
	actual := utils.AbsInt64(input)

	const expected int64 = 10
	if actual != expected {
		t.Errorf("Unexpected value. expected %d got %d", expected, actual)
	}
}
