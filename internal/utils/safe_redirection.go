package utils

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// IsRedirectionSafe determines whether the URL is safe to be redirected to.
func IsRedirectionSafe(url url.URL, protectedDomains []schema.SessionDomainConfiguration) bool {
	if !IsSchemeHTTPS(&url) {
		return false
	}

	if !IsURLUnderProtectedDomain(&url, protectedDomains) {
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
