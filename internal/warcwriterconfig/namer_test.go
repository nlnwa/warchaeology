package warcwriterconfig

import "testing"

func Test_parseSubdirPattern(t *testing.T) {
	tests := []struct {
		name       string
		dirPattern string
		recordDate string
		want       string
		wantErr    bool
	}{
		{"no template", "", "2006-11-17T12:02:28Z", "", false},
		{"4 digit year", "{YYYY}", "2006-11-17T12:02:28Z", "2006", false},
		{"2 digit year", "{YY}", "2006-11-17T12:02:28Z", "06", false},
		{"month", "{MM}", "2006-11-17T12:02:28Z", "11", false},
		{"day", "{DD}", "2006-11-17T12:02:28Z", "17", false},
		{"year month day", "{YYYY}/{MM}/{DD}", "2006-11-17T12:02:28Z", "2006/11/17", false},
		{"extra path", "01/{YYYY}/{MM}/{DD}", "2006-11-17T12:02:28Z", "01/2006/11/17", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSubdirPattern(tt.dirPattern, tt.recordDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSubdirPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSubdirPattern() got = %v, want %v", got, tt.want)
			}
		})
	}
}
