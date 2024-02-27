/*
 * Copyright 2021 National Library of Norway.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package internal

import (
	"net"
	"strings"

	"github.com/nlnwa/whatwg-url/url"
)

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
