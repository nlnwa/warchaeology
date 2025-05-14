package time

import (
	"time"
)

const layout14 = "20060102150405"

// To14 converts a time.Time to a 14 digit string according to the ARC file format.
func To14(t time.Time) string {
	return t.UTC().Format(layout14)
}

// From14ToTime converts a 14 digit string to a time.Time according to the ARC file format.
//
// See https://archive.org/web/researcher/ArcFileFormat.php where date is defined as YYYYMMDDhhmmss (Greenwich Mean Time)
func From14ToTime(s string) (time.Time, error) {
	return time.Parse(layout14, s)
}

// ToW3CDTF returns the time formatted according to W3C Date and Time Formats with up to nanosecond precision.
//
// See https://www.w3.org/TR/NOTE-datetime
func ToW3CDTF(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}
