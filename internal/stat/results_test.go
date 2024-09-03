package stat

import (
	"fmt"
	"testing"
)

const (
	logStringTemplate = "000000 %s: records: %d, processed: %d, errors: %d, duplicates: %d"
	stringTemplate    = "%s: records: %d, processed: %d, errors: %d, duplicates: %d"
)

func TestNewResult(t *testing.T) {
	result := NewResult("test")
	if result == nil {
		t.Errorf("Expected result to be created")
	}
	if result.ErrorCount() != 0 {
		t.Errorf("Expected error count to be 0, got %d", result.ErrorCount())
	}
	if result.Records() != 0 {
		t.Errorf("Expected records to be 0, got %d", result.Records())
	}
	if result.Processed() != 0 {
		t.Errorf("Expected processed to be 0, got %d", result.Processed())
	}
	if result.Duplicates() != 0 {
		t.Errorf("Expected duplicates to be 0, got %d", result.Duplicates())
	}
	if result.Fatal() != nil {
		t.Errorf("Expected fatal to be nil")
	}
}

func TestResultIncrRecords(t *testing.T) {
	result := NewResult("test")
	result.IncrRecords()
	if result.Records() != 1 {
		t.Errorf("Expected records to be 1, got %d", result.Records())
	}
}

func TestResultIncrProcessed(t *testing.T) {
	result := NewResult("test")
	result.IncrProcessed()
	if result.Processed() != 1 {
		t.Errorf("Expected processed to be 1, got %d", result.Processed())
	}
}

func TestResultIncrDuplicates(t *testing.T) {
	result := NewResult("test")
	result.IncrDuplicates()
	if result.Duplicates() != 1 {
		t.Errorf("Expected duplicates to be 1, got %d", result.Duplicates())
	}
}

func TestResultAddError(t *testing.T) {
	result := NewResult("test")
	result.AddError(nil) // This should fail
	if result.ErrorCount() != 1 {
		t.Errorf("Expected error count to be 0, got %d", result.ErrorCount())
	}
}

func TestResultErrors(t *testing.T) {
	result := NewResult("test")
	myError := fmt.Errorf("test error")
	result.AddError(myError)
	if len(result.Errors()) != 1 {
		t.Errorf("Expected errors to be non-empty")
	}
}

func TestResultError(t *testing.T) {
	result := NewResult("test")
	myError := fmt.Errorf("test error")
	result.AddError(myError)
	if result.Error() != "   test error" {
		t.Errorf("Expected error to be 'test error', got %s", result.Error())
	}
}

func TestResultFatal(t *testing.T) {
	result := NewResult("test")
	myError := fmt.Errorf("test error")
	result.SetFatal(myError)
	if result.Fatal() != myError {
		t.Errorf("Expected fatal to be 'test error', got %s", result.Fatal())
	}
}

func TestResultString(t *testing.T) {
	resultName := "test"
	result := NewResult(resultName)
	expectedResultStringTemplate := "%s: records: 0, processed: 0, errors: 0, duplicates: 0"
	expectedResultString := fmt.Sprintf(expectedResultStringTemplate, resultName)
	str := result.String()
	if str != expectedResultString {
		t.Errorf("Expected '%s', got '%s'", expectedResultString, str)

	}
}

func TestResultStringAfterChange(t *testing.T) {
	resultName := "test"
	result := NewResult(resultName)
	expectedErrorCount := 1
	expectedRecords := 0
	expectedProcessed := 0
	expectedDuplicates := 0
	expectedResultString := fmt.Sprintf(stringTemplate, resultName, expectedRecords, expectedProcessed, expectedErrorCount, expectedDuplicates)
	result.AddError(fmt.Errorf("test error"))
	str := result.String()
	if str != expectedResultString {
		t.Errorf("Expected '%s', got '%s'", expectedResultString, str)
	}
}

func TestResultLogWhenEmpty(t *testing.T) {
	result := NewResult("test")
	str := result.Log(0)
	expectedRecords := 0
	expectedProcessed := 0
	expectedErrorCount := 0
	expectedDuplicates := 0
	expectedResultString := fmt.Sprintf(logStringTemplate, "test", expectedRecords, expectedProcessed, expectedErrorCount, expectedDuplicates)
	if str != expectedResultString {
		t.Errorf("Expected '%s', got '%s'", expectedResultString, str)
	}
}

func TestResultLogWhenError(t *testing.T) {
	result := NewResult("test")
	result.AddError(fmt.Errorf("test error"))
	str := result.Log(0)
	expectedRecords := 0
	expectedProcessed := 0
	expectedErrorCount := 1
	expectedDuplicates := 0
	expectedResultString := fmt.Sprintf(logStringTemplate, "test", expectedRecords, expectedProcessed, expectedErrorCount, expectedDuplicates)
	if str != expectedResultString {
		t.Errorf("Expected '%s', got '%s'", expectedResultString, str)
	}
}

func TestResultGetStats(t *testing.T) {
	result := NewResult("test")
	_ = result.GetStats()
}
