package utils

import (
	"fmt"
	"net/url"
	"strings"
)

// IsRedirectionSafe determines whether the URL is safe to be redirected to.
func IsRedirectionSafe(url url.URL, protectedDomain string) bool {
	if url.Scheme != "https" {
		return false
	}

	if url.Hostname() == protectedDomain {
		return true
	}

	if strings.HasSuffix(url.Hostname(), fmt.Sprintf(".%s", protectedDomain)) {
		return true
	}

	return false
}

// IsRedirectionURISafe determines whether the URI is safe to be redirected to.
func IsRedirectionURISafe(uri, protectedDomain string) (bool, error) {
	targetURL, err := url.ParseRequestURI(uri)

	if err != nil {
		return false, fmt.Errorf("Unable to parse redirection URI %s: %w", uri, err)
	}

	return targetURL != nil && IsRedirectionSafe(*targetURL, protectedDomain), nil
}
