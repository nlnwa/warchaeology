package time

import (
	"time"
)

const layout14 = "20060102150405"

func To14(s string) (string, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return "", err
	}
	return t.UTC().Format(layout14), nil
}

// From14ToTime converts a 14 digit string to a time.Time according to the ARC file format.
//
// See https://archive.org/web/researcher/ArcFileFormat.php where date is defined as YYYYMMDDhhmmss (Greenwich Mean Time)
func From14ToTime(s string) (time.Time, error) {
	return time.Parse(layout14, s)
}

// UTCW3CDTF returns the time in UTC formatted according to W3C Date and Time Formats with up to nanosecond precision.
//
// See https://www.w3.org/TR/NOTE-datetime
func UTCW3CDTF(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
