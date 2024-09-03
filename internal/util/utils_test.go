package util_test

import (
	"slices"
	"testing"

	"github.com/nlnwa/warchaeology/v3/internal/util"
)

func TestContainsMatchOnMatching(t *testing.T) {
	needle := "needle"
	haystack := []string{"hay1", "hay2", needle, "hay3", "hay5"}

	if !slices.Contains(haystack, needle) {
		t.Error("Failed to find needle")
	}
}

func TestContainsNonMatchOnNonMatching(t *testing.T) {
	needle := "needle"
	haystack := []string{"hay1", "hay2", "hay4", "hay3", "hay5"}

	if slices.Contains(haystack, needle) {
		t.Error("Found element that does not exist in slice")
	}
}

func TestCropStringCrops(t *testing.T) {
	input := "123456789"
	expected := "1234\u2026"

	actual := util.CropString(input, 5)

	if actual != expected {
		t.Errorf("Expected %s got %s", expected, actual)
	}
}

func TestAbsInt64NegativeInput(t *testing.T) {
	const input int64 = -10
	actual := util.AbsInt64(input)

	const expected int64 = 10
	if actual != expected {
		t.Errorf("Unexpected value. expected %d got %d", expected, actual)
	}
}

func TestAbsInt64PositiveInput(t *testing.T) {
	const input int64 = 10
	actual := util.AbsInt64(input)

	const expected int64 = 10
	if actual != expected {
		t.Errorf("Unexpected value. expected %d got %d", expected, actual)
	}
}
