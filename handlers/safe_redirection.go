package handlers

import (
	"net/url"
	"strings"
)

func isRedirectionSafe(url url.URL, protectedDomain string) bool {
	if url.Scheme != "https" {
		return false
	}

	if !strings.HasSuffix(url.Hostname(), protectedDomain) {
		return false
	}
	return true
}
