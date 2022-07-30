package utils

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func isURLSafe(requestURI string, domains []schema.SessionDomainConfiguration) bool {
	url, _ := url.ParseRequestURI(requestURI)
	return IsRedirectionSafe(*url, domains)
}

func TestIsRedirectionSafe_ShouldReturnFalseOnBadScheme(t *testing.T) {
	assert.False(t, isURLSafe("http://secure.example.com", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
	assert.False(t, isURLSafe("ftp://secure.example.com", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
	assert.True(t, isURLSafe("https://secure.example.com", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
}

func TestIsRedirectionSafe_ShouldReturnFalseOnBadDomain(t *testing.T) {
	assert.False(t, isURLSafe("https://secure.example.com.c", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
	assert.False(t, isURLSafe("https://secure.example.comc", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
	assert.False(t, isURLSafe("https://secure.example.co", []schema.SessionDomainConfiguration{{Domain: "example.com"}}))
}

func TestIsRedirectionURISafe_CannotParseURI(t *testing.T) {
	_, err := IsRedirectionURISafe("http//invalid", []schema.SessionDomainConfiguration{{Domain: "example.com"}})
	assert.EqualError(t, err, "Unable to parse redirection URI http//invalid: parse \"http//invalid\": invalid URI for request")
}

func TestIsRedirectionURISafe_InvalidRedirectionURI(t *testing.T) {
	valid, err := IsRedirectionURISafe("http://myurl.com/myresource", []schema.SessionDomainConfiguration{{Domain: "example.com"}})
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestIsRedirectionURISafe_ValidRedirectionURI(t *testing.T) {
	valid, err := IsRedirectionURISafe("http://myurl.example.com/myresource", []schema.SessionDomainConfiguration{{Domain: "example.com"}})
	assert.NoError(t, err)
	assert.False(t, valid)
}
