package handlers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func isURLSafe(requestURI string, domain string) bool {
	url, _ := url.ParseRequestURI(requestURI)
	return isRedirectionSafe(*url, domain)
}

func TestShouldReturnFalseOnBadScheme(t *testing.T) {
	assert.False(t, isURLSafe("http://secure.example.com", "example.com"))
	assert.False(t, isURLSafe("ftp://secure.example.com", "example.com"))
	assert.True(t, isURLSafe("https://secure.example.com", "example.com"))
}

func TestShouldReturnFalseOnBadDomain(t *testing.T) {
	assert.False(t, isURLSafe("https://secure.example.com.c", "example.com"))
	assert.False(t, isURLSafe("https://secure.example.comc", "example.com"))
	assert.False(t, isURLSafe("https://secure.example.co", "example.com"))
}
