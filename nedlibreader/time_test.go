package nedlibreader

import (
	"testing"
	"time"
)

// CET (Central European Time) aka MET (Middle European Time) aka MEZ (Mitteleuropäische Zeit)
var locCET *time.Location = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		panic(err)
	}
	return loc
}()

const (
	UNIVERSAL                    = "UNIVERSAL"
	RFC1123                      = "RFC1123"
	RFC1123Z                     = "RFC1123Z"
	RFC1123_NO_LEADING_ZERO      = "RFC1123_NO_LEADING_ZERO"
	RFC1123_SECONDS_OUT_OF_RANGE = "RFC1123_SECONDS_OUT_OF_RANGE"
	RFC1123_NORSK                = "RFC1123_NORSK"
	RFC822                       = "RFC822"
	RFC822Z                      = "RFC822Z"
	RFC850                       = "RFC850"
	RFC850_BROKEN_YEAR           = "RFC850_BROKEN_YEAR"
	ANSIC                        = "ANSIC"
	ANSIC_TZ                     = "ANSIC_TZ"
	UnixDate                     = "UnixDate"
	RFC2822                      = "RFC2822"
	RFC2822_TZ                   = "RFC2822_TZ"
)

var timeParseFuncs = map[string]func(string) (time.Time, error){
	RFC1123:                      func(value string) (time.Time, error) { return time.Parse(time.RFC1123, value) },
	RFC1123_NO_LEADING_ZERO:      parseRFC1123NoLeadingZero,
	RFC1123Z:                     func(value string) (time.Time, error) { return time.Parse(time.RFC1123Z, value) },
	RFC1123_NORSK:                parseCustomNorwegianTime,
	RFC1123_SECONDS_OUT_OF_RANGE: parseRFC1123SecondsOutOfRange,
	RFC850:                       func(value string) (time.Time, error) { return time.Parse(time.RFC850, value) },
	RFC850_BROKEN_YEAR:           parseRFC850BrokenYear,
	RFC822:                       func(value string) (time.Time, error) { return time.Parse(time.RFC822, value) },
	RFC822Z:                      func(value string) (time.Time, error) { return time.Parse(time.RFC822Z, value) },
	RFC2822:                      parseRFC2822,
	RFC2822_TZ:                   parseRFC2822TZ,
	ANSIC:                        func(value string) (time.Time, error) { return time.Parse(time.ANSIC, value) },
	ANSIC_TZ:                     parseANSICTZ,
	UnixDate:                     func(value string) (time.Time, error) { return time.Parse(time.UnixDate, value) },
	UNIVERSAL:                    parseTime,
}

var timeTests = map[string][]struct {
	value string
	want  time.Time
}{
	RFC2822: {
		{"23 Jan 2001 19:00:00", time.Date(2001, time.January, 23, 19, 0, 0, 0, time.UTC)},
		{"16 Jul 2003 21:41:40", time.Date(2003, time.July, 16, 21, 41, 40, 0, time.UTC)},
		{"9 Jan 2003 09:23:40", time.Date(2003, time.January, 9, 9, 23, 40, 0, time.UTC)},
	},
	RFC2822_TZ: {
		{"17 Jul 2003 03:53:56 GMT", time.Date(2003, time.July, 17, 3, 53, 56, 0, time.UTC)},
		{"17 Jul 2003 01:53:56 CET", time.Date(2003, time.July, 17, 0, 53, 56, 0, time.UTC)},
	},
	RFC1123_NORSK: {
		{"ma, 04 aug 2003 06:11:40 CET", time.Date(2003, time.August, 4, 7, 11, 40, 0, locCET)},
		{"ti, 29 jul 2003 11:18:55 CET", time.Date(2003, time.July, 29, 12, 18, 55, 0, locCET)},
		{"on, 16 jul 2003 07:41:29 CET", time.Date(2003, time.July, 16, 8, 41, 29, 0, locCET)},
		{"to, 17 jul 2003 01:39:34 CET", time.Date(2003, time.July, 17, 2, 39, 34, 0, locCET)},
		{"fr, 01 aug 2003 05:31:12 GMT", time.Date(2003, time.August, 1, 7, 31, 12, 0, locCET)},
		{"fr, 18 jul 2003 02:10:38 CET", time.Date(2003, time.July, 18, 3, 10, 38, 0, locCET)},
		{"fr, 21 jun 2002 03:53:40 GMT", time.Date(2002, time.June, 21, 5, 53, 40, 0, locCET)},
		{"lø, 19 jul 2003 04:45:41 CET", time.Date(2003, time.July, 19, 5, 45, 41, 0, locCET)},
		{"lø, 02 aug 2003 09:59:08 CET", time.Date(2003, time.August, 2, 10, 59, 8, 0, locCET)},
		{"sø, 03 aug 2003 01:00:43 CET", time.Date(2003, time.August, 3, 2, 0, 43, 0, locCET)},
	},
	ANSIC_TZ: {
		{"Fri Aug  1 23:39:52 2003 CEST", time.Date(2003, time.August, 1, 21, 39, 52, 0, time.UTC)},
	},
	ANSIC: {
		{"Mon Jul 21 05:41:16 2003", time.Date(2003, time.July, 21, 5, 41, 16, 0, time.UTC)},
		{"Tue Jul 22 04:19:03 2003", time.Date(2003, time.July, 22, 4, 19, 3, 0, time.UTC)},
		{"Wed Jul 16 16:58:08 2003", time.Date(2003, time.July, 16, 16, 58, 8, 0, time.UTC)},
		{"Thu Jul 24 00:28:01 2003", time.Date(2003, time.July, 24, 0, 28, 1, 0, time.UTC)},
		{"Fri Jul 18 22:11:59 2003", time.Date(2003, time.July, 18, 22, 11, 59, 0, time.UTC)},
		{"Sat Jul 19 22:09:25 2003", time.Date(2003, time.July, 19, 22, 9, 25, 0, time.UTC)},
		{"Sun Jul 20 20:23:13 2003", time.Date(2003, time.July, 20, 20, 23, 13, 0, time.UTC)},
	},
	RFC1123Z: {
		{"Mon, 04 Aug 2003 21:15:58 +0200", time.Date(2003, time.August, 4, 19, 15, 58, 0, time.UTC)},
		{"Tue, 05 Aug 2003 00:39:50 +0200", time.Date(2003, time.August, 4, 22, 39, 50, 0, time.UTC)},
	},
	RFC1123_NO_LEADING_ZERO: {
		{"Mon, 4 Aug 2003 23:29:21 GMT", time.Date(2003, time.August, 4, 23, 29, 21, 0, time.UTC)},
		{"Tue, 5 Aug 2003 03:40:46 GMT", time.Date(2003, time.August, 5, 3, 40, 46, 0, time.UTC)},
		{"Wed, 6 Aug 2003 21:21:26 GMT", time.Date(2003, time.August, 6, 21, 21, 26, 0, time.UTC)},
		{"Wed, 8 Jan 2003 17:53:27 MET", time.Date(2003, time.January, 8, 18, 53, 27, 0, locCET)},
		{"Thu, 7 Aug 2003 02:18:01 CES", time.Date(2003, time.August, 7, 4, 18, 1, 0, locCET)},
		{"Sat, 2 Aug 2003 03:41:01 GMT", time.Date(2003, time.August, 2, 3, 41, 1, 0, time.UTC)},
		{"Sun, 3 Aug 2003 03:34:00 GMT", time.Date(2003, time.August, 3, 3, 34, 0, 0, time.UTC)},
		{"Sun, 3 Aug 2003 05:43:49 CES", time.Date(2003, time.August, 3, 7, 43, 49, 0, locCET)},
	},
	RFC1123_SECONDS_OUT_OF_RANGE: {
		{"Thu, 17 Jul 2003 12:15:60 GMT", time.Date(2003, time.July, 17, 12, 15, 60, 0, time.UTC)},
		{"Fri, 10 Jan 2003 08:40:60 GMT", time.Date(2003, time.January, 10, 8, 40, 60, 0, time.UTC)},
	},
	RFC850: {
		{"Monday, 04-Aug-03 12:46:57 GMT", time.Date(2003, time.August, 4, 12, 46, 57, 0, time.UTC)},
		{"Tuesday, 05-Aug-03 04:21:56 GMT", time.Date(2003, time.August, 5, 4, 21, 56, 0, time.UTC)},
		{"Wednesday, 06-Aug-03 01:37:20 GMT", time.Date(2003, time.August, 6, 1, 37, 20, 0, time.UTC)},
		{"Wednesday, 16-Jul-03 15:04:47 GMT", time.Date(2003, time.July, 16, 15, 4, 47, 0, time.UTC)},
		{"Thursday, 07-Aug-03 02:33:44 GMT", time.Date(2003, time.August, 7, 2, 33, 44, 0, time.UTC)},
		{"Friday, 18-Jul-03 14:18:04 GMT", time.Date(2003, time.July, 18, 14, 18, 4, 0, time.UTC)},
		{"Saturday, 26-Jul-03 00:44:32 GMT", time.Date(2003, time.July, 26, 0, 44, 32, 0, time.UTC)},
		{"Sunday, 20-Jul-03 20:24:12 GMT", time.Date(2003, time.July, 20, 20, 24, 12, 0, time.UTC)},
	},
	RFC850_BROKEN_YEAR: {
		// This is obviously wrong, but it's what the original record contains (assume year 2003)
		{"Wednesday, 24-Jul-:3 03:33:15 GMT", time.Date(2003, time.July, 24, 3, 33, 15, 0, time.UTC)},
		// This is obviously wrong, but it's what the original record contains (assume year 2003)
		{"Friday, 18-Jul-103 09:39:49 GMT", time.Date(2003, time.July, 18, 9, 39, 49, 0, time.UTC)},
		{"Monday, 21-Jul-103 01:39:08 GMT", time.Date(2003, time.July, 21, 1, 39, 8, 0, time.UTC)},
	},
}

func TestParseTime(t *testing.T) {
	// set location to Paris for consistent time parsing in tests
	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Fatal(err)
	}
	time.Local = location

	for testFormat, tests := range timeTests {
		_, ok := timeParseFuncs[testFormat]
		if !ok {
			t.Errorf("parsing of '%s' is not implemented", testFormat)
			continue
		}
		for parseFormat, fn := range timeParseFuncs {
			for _, test := range tests {
				t.Run(testFormat+"-"+test.value, func(t *testing.T) {
					// parse the time
					got, err := fn(test.value)

					// expect error when parsing a different format
					expectError := parseFormat != testFormat
					if parseFormat == UNIVERSAL {
						// universal function should parse all formats
						expectError = false
					}
					// assert no error when not expected
					if !expectError && err != nil {
						t.Fatalf("(parsed as '%s') unexpected error: %v", parseFormat, err)
					}
					// assert time is parsed correctly
					if !expectError && !got.Equal(test.want) {
						t.Errorf("(parsed as '%s') want: '%v', got: '%v': ", parseFormat, test.want, got)
					}
				})
			}
		}
	}
}
