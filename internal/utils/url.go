package utils

import (
	"fmt"
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

// IsURIStringSafeRedirection determines whether the URI is safe to be redirected to.
func IsURIStringSafeRedirection(uri, protectedDomain string) (safe bool, err error) {
	var parsedURI *url.URL

	if parsedURI, err = url.ParseRequestURI(uri); err != nil {
		return false, fmt.Errorf("failed to parse URI '%s': %w", uri, err)
	}

	return parsedURI != nil && IsURISafeRedirection(parsedURI, protectedDomain), nil
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

	if strings.HasSuffix(domain, period+domainSuffix) {
		return true
	}

	return false
}
