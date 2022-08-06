package utils

import (
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// IsURLUnderProtectedDomain returns true if specied url is under Authelia's protected domains.
func IsURLUnderProtectedDomain(url *url.URL, domains []schema.SessionDomainConfiguration) (bool, string) {
	hostname := url.Hostname()
	for _, domain := range domains {
		if strings.HasSuffix(hostname, domain.Domain) && domain.Domain != "" {
			return true, domain.Domain
		}
	}

	return false, ""
}

// IsSchemeHTTPS return true if url scheme is https.
func IsSchemeHTTPS(url *url.URL) bool {
	return url.Scheme == "https"
}
