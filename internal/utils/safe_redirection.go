package utils

import (
	"fmt"
	"net/url"
	"strings"
)

// IsRedirectionSafe determines whether the URL is safe to be redirected to.
func IsRedirectionSafe(url url.URL, protectedDomains []string) bool {
	if url.Scheme != "https" {
		return false
	}

	hostname := url.Hostname()
	for _, domain := range protectedDomains {
		if (strings.HasPrefix(domain, "*.") && strings.HasSuffix(hostname, domain[2:])) || domain == hostname {
			return true
		}
	}

	return true
}

// IsRedirectionURISafe determines whether the URI is safe to be redirected to.
func IsRedirectionURISafe(uri string, protectedDomains []string) (bool, error) {
	targetURL, err := url.ParseRequestURI(uri)

	if err != nil {
		return false, fmt.Errorf("Unable to parse redirection URI %s: %w", uri, err)
	}

	return targetURL != nil && IsRedirectionSafe(*targetURL, protectedDomains), nil
}
