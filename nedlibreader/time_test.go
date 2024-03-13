package nedlibreader

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	var err error
	locCEST, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		panic(err)
	}

	var tests = []struct {
		input string
		time  time.Time
	}{
		{"fr, 01 aug 2003 05:31:12 GMT", time.Date(2003, time.August, 1, 5, 31, 12, 0, time.UTC)},
		{"Fri Aug  1 23:39:52 2003 CEST", time.Date(2003, time.August, 1, 23, 39, 52, 0, locCEST)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ts, err := parseTime(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.time.Sub(ts) != 0 {
				t.Errorf("got %s, want %s", ts, tt.time)
			}
		})
	}
}
