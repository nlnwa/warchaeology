package nedlibreader

import (
	"testing"
	"time"
)

type testTime struct {
	timeAsString string
	time         time.Time
	shouldFail   bool
}

type testTable struct {
	expectedFormat string
	tests          []testTime
}

func TestParseTime(t *testing.T) {
	var testTable = []testTable{
		{
			expectedFormat: RFC1123,
			tests: []testTime{
				{"Tue, 05 Apr 2024 15:30:00 GMT", time.Date(2024, time.April, 5, 15, 30, 0, 0, time.UTC), false},
				{"Tue, 05 Apr 202 15:30:00 GMT", time.Date(2024, time.April, 5, 15, 30, 0, 0, time.UTC), true},
			},
		},
		{
			expectedFormat: RFC850,
			tests: []testTime{
				{"Tuesday, 05-Apr-24 15:30:00 GMT", time.Date(2024, time.April, 5, 15, 30, 0, 0, time.UTC), false},
				{"Tue, 05 Ap 2024 15:30:00 GMT", time.Date(2024, time.April, 5, 15, 30, 0, 0, time.UTC), true},
			},
		},
		{
			expectedFormat: ANSIC,
			tests: []testTime{
				{"Tue Apr  5 15:30:00 2024", time.Date(2024, time.April, 5, 15, 30, 0, 0, time.UTC), false},
				{"Tue, 05 Apr 204 15:30:00 GMT", time.Date(2024, time.April, 5, 15, 30, 0, 0, time.UTC), true},
			},
		},
		{
			expectedFormat: UnixDate,
			tests: []testTime{
				{"Tue Apr  5 15:30:00 UTC 2024", time.Date(2024, time.April, 5, 15, 30, 0, 0, time.UTC), false},
				{"Tue, 5 Apr 2024 15:30:00 GMT", time.Date(2024, time.April, 5, 15, 30, 0, 0, time.UTC), true},
			},
		},
	}

	for _, formatGroup := range testTable {
		for _, test := range formatGroup.tests {
			t.Run(formatGroup.expectedFormat+"-"+test.timeAsString, func(t *testing.T) {
				t.Log("Time string: ", test.timeAsString)
				t.Log("Expected time: ", test.time)
				t.Log("Should fail: ", test.shouldFail)
				t.Log("Expected format: ", formatGroup.expectedFormat)
				parsedTime, detectedFormat, err := parseTime(test.timeAsString)

				if err != nil {
					if !test.shouldFail {
						t.Errorf("expected no error, got: %v", err)
					}
				} else {
					if test.shouldFail {
						t.Errorf("expected error, got none")
					}
					if detectedFormat != formatGroup.expectedFormat {
						t.Errorf("expected format: '%v', got: '%v'", formatGroup.expectedFormat, detectedFormat)
					}
					if !parsedTime.Equal(test.time) {
						t.Errorf("expected time: '%v', got: '%v'", test.time, parsedTime)
					}
				}
			})
		}
	}
}
