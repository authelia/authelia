package utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

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
