package utils

import (
	"net/url"
	"strings"
)

func IsRedirectionSafe(url url.URL, protectedDomain string) bool {
	if url.Scheme != "https" {
		return false
	}

	if !strings.HasSuffix(url.Hostname(), protectedDomain) {
		return false
	}
	return true
}
