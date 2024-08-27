package ssurt

import (
	"net"
	"strings"

	whatwgUrl "github.com/nlnwa/whatwg-url/url"
)

// SSURT stands for Superior SURT. Sensible SURT. Smug SURT.
// SURT stands for Sort-friendly URI Reordering Transform
// For more information, see https://github.com/iipc/urlcanon/blob/master/ssurt.rst

func SsurtUrl(url *whatwgUrl.Url, includeScheme bool) (string, error) {
	url.SearchParams().Sort()

	var result strings.Builder
	hostname := url.Hostname()
	if hostname != "" {
		if hostname[0] == '[' {
			result.WriteString(hostname)
		} else if net.ParseIP(hostname).To4() != nil {
			result.WriteString(hostname)
		} else {
			splitOnDot := strings.Split(hostname, ".")
			for partIndex := len(splitOnDot) - 1; partIndex >= 0; partIndex-- {
				result.WriteString(splitOnDot[partIndex])
				result.WriteByte(',')
			}
		}
		result.WriteString("//")
	}
	if includeScheme {
		if url.Port() != "" {
			result.WriteString(url.Port())
			result.WriteByte(':')
		}
		result.WriteString(strings.TrimSuffix(url.Protocol(), ":"))
		if url.Username() != "" {
			result.WriteByte('@')
			result.WriteString(url.Username())
		}
		if url.Password() != "" {
			result.WriteByte(':')
			result.WriteString(url.Password())
		}
		result.WriteByte(':')
	}
	result.WriteString(url.Pathname())
	result.WriteString(url.Search())
	result.WriteString(url.Hash())

	return result.String(), nil
}

var parser = whatwgUrl.NewParser(whatwgUrl.WithSkipEqualsForEmptySearchParamsValue())

func SsurtString(urlAsString string, includeScheme bool) (string, error) {
	url, err := parser.Parse(urlAsString)
	if err != nil {
		return "", err
	}
	return SsurtUrl(url, includeScheme)
}
