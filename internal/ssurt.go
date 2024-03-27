package internal

import (
	"net"
	"strings"

	"github.com/nlnwa/whatwg-url/url"
)

// SSURT stands for Superior SURT. Sensible SURT. Smug SURT.
// SURT stands for Sort-friendly URI Reordering Transform
// For more information, see https://github.com/iipc/urlcanon/blob/master/ssurt.rst

func SsurtUrl(u *url.Url, includeScheme bool) (string, error) {
	u.SearchParams().Sort()

	var result strings.Builder
	hostname := u.Hostname()
	if hostname != "" {
		if hostname[0] == '[' {
			result.WriteString(hostname)
		} else if net.ParseIP(hostname).To4() != nil {
			result.WriteString(hostname)
		} else {
			t := strings.Split(hostname, ".")
			for i := len(t) - 1; i >= 0; i-- {
				result.WriteString(t[i])
				result.WriteByte(',')
			}
		}
		result.WriteString("//")
	}
	if includeScheme {
		if u.Port() != "" {
			result.WriteString(u.Port())
			result.WriteByte(':')
		}
		result.WriteString(strings.TrimSuffix(u.Protocol(), ":"))
		if u.Username() != "" {
			result.WriteByte('@')
			result.WriteString(u.Username())
		}
		if u.Password() != "" {
			result.WriteByte(':')
			result.WriteString(u.Password())
		}
		result.WriteByte(':')
	}
	result.WriteString(u.Pathname())
	result.WriteString(u.Search())
	result.WriteString(u.Hash())

	return result.String(), nil
}

var surtParser = url.NewParser(url.WithSkipEqualsForEmptySearchParamsValue())

func SsurtString(u string, includeScheme bool) (string, error) {
	u2, err := surtParser.Parse(u)
	if err != nil {
		return "", err
	}
	return SsurtUrl(u2, includeScheme)
}
