// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"net"
	"net/url"
	"path"
	"strings"
)

// URIAccessPath returns a URL path with the query parameters appended (full path) with the path portion parsed
// through [path.Clean] after dot segments are removed per RFC 3986 section 5.2.4 given a *[url.URL]. An empty path is
// treated as '/' per RFC 3986 section 6.2.3 and RFC 9110 section 4.2.3.
func URIAccessPath(u *url.URL) (output string) {
	p := u.Path

	if p == "" {
		p = "/"
	}

	output = path.Clean(urlRemoveDotSegments(p))

	if output != "/" && len(p) > 1 && p[len(p)-1] == '/' {
		output += "/"
	}

	if len(u.RawQuery) != 0 {
		output += "?" + u.RawQuery
	}

	return output
}

// IsURISecure returns true if the URI has a secure schemes (https or wss).
func IsURISecure(uri *url.URL) bool {
	switch uri.Scheme {
	case https, wss:
		return true
	default:
		return false
	}
}

// HasURIDomainSuffix returns true if the URI hostname is equal to the domain suffix or if it has a suffix of the domain
// suffix prefixed with a period.
func HasURIDomainSuffix(uri *url.URL, domainSuffix string) bool {
	return HasDomainSuffix(uri.Hostname(), domainSuffix)
}

// HasDomainSuffix returns true if the URI hostname is equal to the domain or if it has a suffix of the domain
// prefixed with a period.
func HasDomainSuffix(domain, domainSuffix string) bool {
	if domainSuffix == "" {
		return false
	}

	if strings.EqualFold(domain, domainSuffix) {
		return true
	}

	if (strings.HasPrefix(domainSuffix, period) && StringHasSuffixFold(domain, domainSuffix)) || StringHasSuffixFold(domain, period+domainSuffix) {
		return true
	}

	return false
}

// IsHostnameValid returns true if the hostname is either an IP address or a valid DNS hostname consisting of labels of
// 1 to 63 alphanumeric or hyphen characters which do not start or end with a hyphen, with a total length of no more
// than 253 characters.
func IsHostnameValid(hostname string) (valid bool) {
	if net.ParseIP(hostname) != nil {
		return true
	}

	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	for label := range strings.SplitSeq(hostname, period) {
		if !isHostnameLabelValid(label) {
			return false
		}
	}

	return true
}

func isHostnameLabelValid(label string) (valid bool) {
	n := len(label)

	if n == 0 || n > 63 || label[0] == '-' || label[n-1] == '-' {
		return false
	}

	for i := range n {
		switch c := label[i]; {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-':
			continue
		default:
			return false
		}
	}

	return true
}

// EqualURLs returns true if the two *[url.URL] values are effectively equal taking into consideration web normalization.
func EqualURLs(first, second *url.URL) bool {
	if first == nil && second == nil {
		return true
	} else if first == nil || second == nil {
		return false
	}

	if !strings.EqualFold(first.Scheme, second.Scheme) {
		return false
	}

	if !strings.EqualFold(first.Host, second.Host) {
		return false
	}

	if first.Path != second.Path {
		return false
	}

	if first.RawQuery != second.RawQuery {
		return false
	}

	if first.Fragment != second.Fragment {
		return false
	}

	if first.RawFragment != second.RawFragment {
		return false
	}

	return true
}

// IsURLInSlice returns true if the needle [url.URL] is in the []url.URL haystack.
func IsURLInSlice(needle *url.URL, haystack []*url.URL) (has bool) {
	for i := 0; i < len(haystack); i++ {
		if EqualURLs(needle, haystack[i]) {
			return true
		}
	}

	return false
}
