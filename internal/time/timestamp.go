package time

import (
	"time"
)

func To14(s string) (string, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return "", err
	}
	return t.UTC().Format("20060102150405"), nil
}

func FromTimeTo14(t time.Time) string {
	return t.Format("20060102150405")
}

func From14ToTime(s string) (time.Time, error) {
	t, err := time.Parse("20060102150405", s)
	return t, err
}

func UTC(t time.Time) time.Time {
	return t.In(time.UTC)
}

func UTC14(t time.Time) string {
	return t.In(time.UTC).Format("20060102150405")
}

// UTCW3CDTF returns the time in UTC formatted according to W3C Date and Time Formats with up to nanosecond precision.
//
// See https://www.w3.org/TR/NOTE-datetime
func UTCW3CDTF(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
