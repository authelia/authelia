package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainMatcher(t *testing.T) {
	assert.True(t, isDomainMatching("example.com", "example.com"))

	assert.False(t, isDomainMatching("example.com", "*.example.com"))
	assert.True(t, isDomainMatching("abc.example.com", "*.example.com"))
	assert.True(t, isDomainMatching("abc.def.example.com", "*.example.com"))

	// Character * must be followed by . to be valid.
	assert.False(t, isDomainMatching("example.com", "*example.com"))

	assert.False(t, isDomainMatching("example.com", "*.example.com"))
	assert.False(t, isDomainMatching("example.com", "*.exampl.com"))

	assert.False(t, isDomainMatching("example.com", "*.other.net"))
	assert.False(t, isDomainMatching("example.com", "*other.net"))
	assert.False(t, isDomainMatching("example.com", "other.net"))
}
