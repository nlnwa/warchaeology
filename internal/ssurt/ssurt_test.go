package ssurt

import (
	"testing"
)

func TestSsurtS(t *testing.T) {
	tests := []struct {
		name          string
		u             string
		includeScheme bool
		want          string
		wantErr       bool
	}{
		{"1", "http://www.example.com", true, "com,example,www,//http:/", false},
		{"2", "http://www.example.com:80", true, "com,example,www,//http:/", false},
		{"3", "http://www.example.com/foo/bar", true, "com,example,www,//http:/foo/bar", false},
		{"4", "http://127.0.0.1/foo/bar", true, "127.0.0.1//http:/foo/bar", false},
		{"5", "http://[::1]/foo/bar", true, "[::1]//http:/foo/bar", false},
		{"11", "http://example.com/foo/bar?query#fragment", true, "com,example,//http:/foo/bar?query#fragment", false},
		{"12", "http://example.com:8080/foo/bar?query#fragment", true, "com,example,//8080:http:/foo/bar?query#fragment", false},
		{"13", "http://user:pass@foo.example.org:81/path?query#frag", true, "org,example,foo,//81:http@user:pass:/path?query#frag", false},
		{"14", "http://user@foo.example.org:81/path?query#frag", true, "org,example,foo,//81:http@user:/path?query#frag", false},
		{"15", "http://foo.example.org:81/path?query#frag", true, "org,example,foo,//81:http:/path?query#frag", false},
		{"16", "http://foo.example.org/path?query#frag", true, "org,example,foo,//http:/path?query#frag", false},
		{"17", "http://81.foo.example.org/path?query#frag", true, "org,example,foo,81,//http:/path?query#frag", false},
		{"18", "screenshot:http://example.com/", true, "screenshot:http://example.com/", false},
		{"19", "scheme://user:pass@foo.example.org:81/path?query#frag", true, "org,example,foo,//81:scheme@user:pass:/path?query#frag", false},
		{"20", "scheme://user@foo.example.org:81/path?query#frag", true, "org,example,foo,//81:scheme@user:/path?query#frag", false},
		{"21", "scheme://foo.example.org:81/path?query#frag", true, "org,example,foo,//81:scheme:/path?query#frag", false},
		{"22", "scheme://foo.example.org/path?query#frag", true, "org,example,foo,//scheme:/path?query#frag", false},
		{"23", "scheme://81.foo.example.org/path?query#frag", true, "org,example,foo,81,//scheme:/path?query#frag", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SsurtString(tt.u, tt.includeScheme)
			if (err != nil) != tt.wantErr {
				t.Errorf("SsurtS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SsurtS() got = %v, want %v", got, tt.want)
			}
		})
	}
}
