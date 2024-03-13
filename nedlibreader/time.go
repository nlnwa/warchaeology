package nedlibreader

import (
	"fmt"
	"time"
)

func parseTime(dateString string) (t time.Time, err error) {
	t, err = time.Parse(time.RFC1123, dateString)
	if err == nil {
		return
	}

	t, err = time.Parse(time.RFC850, dateString)
	if err == nil {
		return
	}

	err = fmt.Errorf("failed to parse string as time.Time: %s", dateString)
	return
}
