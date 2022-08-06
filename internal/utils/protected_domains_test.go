package utils

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func isProtectedURL(url *url.URL, domains []schema.SessionDomainConfiguration) bool {
	protected, _ := IsURLUnderProtectedDomain(url, domains)
	return protected
}

func TestIsDomainProtected(t *testing.T) {
	GetURL := func(u string) *url.URL {
		x, err := url.ParseRequestURI(u)
		require.NoError(t, err)

		return x
	}

	domains := []schema.SessionDomainConfiguration{
		{
			Domain: "example.com",
		},
		{
			Domain: "example2.com",
		},
	}

	assert.True(t, isProtectedURL(
		GetURL("http://mytest.example.com/abc/?query=abc"), domains))

	assert.True(t, isProtectedURL(
		GetURL("http://example.com/abc/?query=abc"), domains))

	assert.True(t, isProtectedURL(
		GetURL("https://mytest.example.com/abc/?query=abc"), domains))

	// Cookies readable by a service on a machine is also readable by a service on the same machine
	// with a different port as mentioned in https://tools.ietf.org/html/rfc6265#section-8.5.
	assert.True(t, isProtectedURL(
		GetURL("https://mytest.example.com:8080/abc/?query=abc"), domains))

	assert.True(t, isProtectedURL(
		GetURL("https://mytest.example2.com/abc/?query=abc"), domains))

	assert.False(t, isProtectedURL(
		GetURL("https://mytest.example3.com/abc/?query=abc"), domains))
}

func TestSchemeIsHTTPS(t *testing.T) {
	GetURL := func(u string) *url.URL {
		x, err := url.ParseRequestURI(u)
		require.NoError(t, err)

		return x
	}

	assert.False(t, IsSchemeHTTPS(
		GetURL("http://mytest.example.com/abc/?query=abc")))
	assert.False(t, IsSchemeHTTPS(
		GetURL("ws://mytest.example.com/abc/?query=abc")))
	assert.False(t, IsSchemeHTTPS(
		GetURL("wss://mytest.example.com/abc/?query=abc")))
	assert.True(t, IsSchemeHTTPS(
		GetURL("https://mytest.example.com/abc/?query=abc")))
}
