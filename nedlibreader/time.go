package nedlibreader

import (
	"fmt"
	"time"
)

const (
	RFC1123  = "RFC1123"
	RFC850   = "RFC850"
	ANSIC    = "ANSIC"
	UnixDate = "UnixDate"
)

func parseTime(dateString string) (time.Time, string, error) {
	parsedTime, err := time.Parse(time.RFC1123, dateString)
	if err == nil {
		return parsedTime, RFC1123, err
	}

	parsedTime, err = time.Parse(time.RFC850, dateString)
	if err == nil {
		return parsedTime, RFC850, err
	}

	parsedTime, err = time.Parse(time.ANSIC, dateString)
	if err == nil {
		return parsedTime, ANSIC, err
	}

	parsedTime, err = time.Parse(time.UnixDate, dateString)
	if err == nil {
		return parsedTime, UnixDate, err
	}

	err = fmt.Errorf("failed to parse string as time.Time: '%s': '%w'", dateString, err)
	return time.Time{}, "", err
}
