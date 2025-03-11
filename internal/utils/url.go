package utils

import (
	"net/url"
	"path"
	"strings"
)

// URLPathFullClean returns a URL path with the query parameters appended (full path) with the path portion parsed
// through path.Clean given a *url.URL.
func URLPathFullClean(u *url.URL) (output string) {
	lengthPath := len(u.Path)
	lengthQuery := len(u.RawQuery)
	appendForwardSlash := lengthPath > 1 && u.Path[lengthPath-1] == '/'

	switch {
	case lengthPath == 1 && lengthQuery == 0:
		return u.Path
	case lengthPath == 1:
		return path.Clean(u.Path) + "?" + u.RawQuery
	case lengthQuery != 0 && appendForwardSlash:
		return path.Clean(u.Path) + "/?" + u.RawQuery
	case lengthQuery != 0:
		return path.Clean(u.Path) + "?" + u.RawQuery
	case appendForwardSlash:
		return path.Clean(u.Path) + "/"
	default:
		return path.Clean(u.Path)
	}
}

// IsURISafeRedirection returns true if the URI passes the IsURISecure and HasURIDomainSuffix, i.e. if the scheme is
// secure and the given URI has a hostname that is either exactly equal to the given domain or if it has a suffix of the
// domain prefixed with a period.
func IsURISafeRedirection(uri *url.URL, domain string) bool {
	return IsURISecure(uri) && HasURIDomainSuffix(uri, domain)
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

	if domain == domainSuffix {
		return true
	}

	if (strings.HasPrefix(domainSuffix, period) && strings.HasSuffix(domain, domainSuffix)) || strings.HasSuffix(domain, period+domainSuffix) {
		return true
	}

	return false
}

// EqualURLs returns true if the two *url.URL values are effectively equal taking into consideration web normalization.
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

// IsURLInSlice returns true if the needle url.URL is in the []url.URL haystack.
func IsURLInSlice(needle *url.URL, haystack []*url.URL) (has bool) {
	for i := 0; i < len(haystack); i++ {
		if EqualURLs(needle, haystack[i]) {
			return true
		}
	}

	return false
}
