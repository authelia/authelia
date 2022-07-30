package utils

import (
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// IsURLUnderProtectedDomain returns true if specied url is under Authelia's protected domains.
func IsURLUnderProtectedDomain(url *url.URL, domains []schema.SessionDomainConfiguration) bool {
	hostname := url.Hostname()
	for _, domain := range domains {
		if strings.HasSuffix(hostname, domain.Domain) {
			return true
		}
	}

	return false
}

// IsSchemeHTTPS return true if url scheme is https.
func IsSchemeHTTPS(url *url.URL) bool {
	return url.Scheme == "https"
}
