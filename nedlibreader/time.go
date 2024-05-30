package nedlibreader

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	// embed timezone data in binary by adding -tags timetzdata to the go build
	// command or uncomment the line below: (see https://pkg.go.dev/time/tzdata)
	// _ "time/tzdata"
)

func parseTime(dateString string) (t time.Time, err error) {
	t, err = time.Parse(time.RFC1123, dateString)
	if err == nil {
		return
	}
	t, err = time.Parse(time.RFC1123Z, dateString)
	if err == nil {
		return
	}
	t, err = time.Parse(time.RFC850, dateString)
	if err == nil {
		return
	}
	t, err = time.Parse(time.ANSIC, dateString)
	if err == nil {
		return
	}
	t, err = time.Parse(time.UnixDate, dateString)
	if err == nil {
		return
	}
	t, err = time.Parse(time.RFC822, dateString)
	if err == nil {
		return
	}
	t, err = time.Parse(time.RFC822Z, dateString)
	if err == nil {
		return
	}
	t, err = parseRFC1123NoLeadingZero(dateString)
	if err == nil {
		return
	}
	t, err = parseRFC1123SecondsOutOfRange(dateString)
	if err == nil {
		return
	}
	t, err = parseRFC2822(dateString)
	if err == nil {
		return
	}
	t, err = parseRFC2822TZ(dateString)
	if err == nil {
		return
	}
	t, err = parseANSICTZ(dateString)
	if err == nil {
		return
	}
	t, err = parseCustomNorwegianTime(dateString)
	if err == nil {
		return
	}
	t, err = parseRFC850BrokenYear(dateString)
	if err == nil {
		return
	}
	t, err = parseRFC1123BrokenYear(dateString)
	if err == nil {
		return
	}
	return time.Time{}, fmt.Errorf("failed to parse string as time.Time: '%s'", dateString)
}

func parseRFC1123NoLeadingZero(value string) (time.Time, error) {
	return time.Parse("Mon, 2 Jan 2006 15:04:05 MST", value)
}

func parseRFC1123SecondsOutOfRange(value string) (t time.Time, err error) {
	// Replace "60" with "59"
	value = regexp.MustCompile(`:60 `).ReplaceAllLiteralString(value, ":59 ")
	t, err = time.Parse(time.RFC1123, value)
	return t.Add(time.Second), err
}

func parseRFC2822(value string) (time.Time, error) {
	return time.Parse("2 Jan 2006 15:04:05", value)
}

func parseRFC2822TZ(value string) (time.Time, error) {
	return time.Parse("2 Jan 2006 15:04:05 MST", value)
}

func parseANSICTZ(value string) (time.Time, error) {
	return time.Parse("Mon Jan _2 15:04:05 2006 MST", value)
}

func parseRFC850BrokenYear(value string) (time.Time, error) {
	// Replace Jul-:3 With Jul-03
	value = regexp.MustCompile(`(\d{2}-[A-Za-z]{3}-):(\d)`).ReplaceAllString(value, "${1}0${2}")
	// Replace Jul-103 With Jul-03
	value = regexp.MustCompile(`([A-Za-z]{3}-)1(\d{2})`).ReplaceAllString(value, "${1}${2}")
	return time.Parse(time.RFC850, value)
}

func parseRFC1123BrokenYear(value string) (time.Time, error) {
	// Replace 31 Jul 103 With 31 Jul 2003
	value = regexp.MustCompile(`(\d{2} [A-Za-z]{3}) 1(\d{2})`).ReplaceAllString(value, "${1} 20${2}")
	return time.Parse(time.RFC1123, value)
}

// parse date and time string on the format: "lø, 19 jul 2003 04:45:41 CET" to a time.Time
func parseCustomNorwegianTime(dateString string) (t time.Time, err error) {
	days := map[string]time.Weekday{
		"ma": time.Monday,
		"ti": time.Tuesday,
		"on": time.Wednesday,
		"to": time.Thursday,
		"fr": time.Friday,
		"lø": time.Saturday,
		"sø": time.Sunday,
	}
	months := map[string]time.Month{
		"jan": time.January,
		"feb": time.February,
		"mar": time.March,
		"apr": time.April,
		"mai": time.May,
		"jun": time.June,
		"jul": time.July,
		"aug": time.August,
		"sep": time.September,
		"okt": time.October,
		"nov": time.November,
		"des": time.December,
	}

	parts := strings.Split(dateString, " ")
	if len(parts) != 6 {
		err = fmt.Errorf("failed to parse date and time: %s", dateString)
		return
	}

	dayPart := strings.TrimSuffix(parts[0], ",")
	dayOfMonthPart := parts[1]
	monthPart := parts[2]
	yearPart := parts[3]
	timePart := parts[4]
	timeZonePart := parts[5]

	_, ok := days[dayPart]
	if !ok {
		err = fmt.Errorf("failed to parse day of the week: %s", dayPart)
		return
	}

	// parse the day of the month
	dayOfMonth, err := strconv.Atoi(dayOfMonthPart)
	if err != nil {
		err = fmt.Errorf("failed to parse day of the month: %s", dayOfMonthPart)
		return
	}

	// parse the month
	month, ok := months[monthPart]
	if !ok {
		err = fmt.Errorf("failed to parse month: %s", monthPart)
		return
	}
	// parse the year
	year, err := strconv.Atoi(yearPart)
	if err != nil {
		err = fmt.Errorf("failed to parse year: %s", yearPart)
		return
	}
	// parse the time
	timeParts := strings.Split(timePart, ":")
	if len(timeParts) != 3 {
		err = fmt.Errorf("failed to parse time: %s", timePart)
		return
	}
	var hour int
	hour, err = strconv.Atoi(timeParts[0])
	if err != nil {
		err = fmt.Errorf("failed to parse hour: %s", timeParts[0])
		return
	}
	var minute int
	minute, err = strconv.Atoi(timeParts[1])
	if err != nil {
		err = fmt.Errorf("failed to parse minute: %s", timeParts[1])
		return
	}
	var second int
	second, err = strconv.Atoi(timeParts[2])
	if err != nil {
		err = fmt.Errorf("failed to parse second: %s", timeParts[2])
		return
	}
	// Assume location is Europe/Oslo
	location, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		err = fmt.Errorf("failed to load location: %w", err)
		return
	}
	// Parse time as observed from Europe/Oslo
	t = time.Date(year, month, dayOfMonth, hour, minute, second, 0, time.UTC)
	rfc1123 := t.Format("Mon, 02 Jan 2006 15:04:05") + " " + timeZonePart
	return time.ParseInLocation(time.RFC1123, rfc1123, location)
}
