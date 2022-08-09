package utils

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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

// URLDomainHasSuffix determines whether the uri has a suffix of the domain value.
func URLDomainHasSuffix(uri url.URL, domain string) bool {
	if uri.Scheme != https {
		return false
	}

	if uri.Hostname() == domain {
		return true
	}

	if strings.HasSuffix(uri.Hostname(), period+domain) {
		return true
	}

	return false
}

// IsRedirectionSafe determines whether the URL is safe to be redirected to.
func IsRedirectionSafe(url url.URL, protectedDomains []schema.SessionDomainConfiguration) bool {
	if !IsSchemeHTTPS(&url) {
		return false
	}

	if protected, _ := IsURLUnderProtectedDomain(&url, protectedDomains); !protected {
		return false
	}

	return true
}

// IsRedirectionURISafe determines whether the URI is safe to be redirected to.
func IsRedirectionURISafe(uri string, protectedDomains []schema.SessionDomainConfiguration) (bool, error) {
	targetURL, err := url.ParseRequestURI(uri)

	if err != nil {
		return false, fmt.Errorf("Unable to parse redirection URI %s: %w", uri, err)
	}

	return targetURL != nil && IsRedirectionSafe(*targetURL, protectedDomains), nil
}

// GetPortalURL gets redirection URL from session configuration.
func GetPortalURL(uri string, domains []schema.SessionDomainConfiguration) string {
	targetURL, err := url.ParseRequestURI(uri)
	if err != nil {
		return ""
	}

	hostname := targetURL.Hostname()

	for _, domain := range domains {
		if strings.HasSuffix(hostname, domain.Domain) {
			return domain.PortalURL
		}
	}

	return ""
}
