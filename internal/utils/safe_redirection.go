package utils

import (
	"net/url"
	"strings"
)

// IsRedirectionSafe determines if a redirection URL is secured.
func IsRedirectionSafe(url url.URL, protectedDomain string) bool {
	if url.Scheme != "https" {
		return false
	}

	if !strings.HasSuffix(url.Hostname(), protectedDomain) {
		return false
	}
	return true
}
