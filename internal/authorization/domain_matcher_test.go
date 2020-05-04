package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldMatchACLWithSingleDomain(t *testing.T) {
	assert.True(t, isDomainMatching("example.com", []string{"example.com"}))

	assert.True(t, isDomainMatching("abc.example.com", []string{"*.example.com"}))
	assert.True(t, isDomainMatching("abc.def.example.com", []string{"*.example.com"}))
}

func TestShouldNotMatchACLWithSingleDomain(t *testing.T) {
	assert.False(t, isDomainMatching("example.com", []string{"*.example.com"}))
	// Character * must be followed by . to be valid.
	assert.False(t, isDomainMatching("example.com", []string{"*example.com"}))

	assert.False(t, isDomainMatching("example.com", []string{"*.exampl.com"}))

	assert.False(t, isDomainMatching("example.com", []string{"*.other.net"}))
	assert.False(t, isDomainMatching("example.com", []string{"*other.net"}))
	assert.False(t, isDomainMatching("example.com", []string{"other.net"}))
}

func TestShouldMatchACLWithMultipleDomains(t *testing.T) {
	assert.True(t, isDomainMatching("example.com", []string{"*.example.com", "example.com"}))
	assert.True(t, isDomainMatching("apple.example.com", []string{"*.example.com", "example.com"}))
}

func TestShouldNotMatchACLWithMultipleDomains(t *testing.T) {
	assert.False(t, isDomainMatching("example.com", []string{"*.example.com", "*example.com"}))
	assert.False(t, isDomainMatching("apple.example.com", []string{"*example.com", "example.com"}))
}
