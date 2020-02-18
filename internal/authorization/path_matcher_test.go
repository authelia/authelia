package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathMatcher(t *testing.T) {
	// Matching any path if no regexp is provided
	assert.True(t, isPathMatching("/", []string{}))

	assert.False(t, isPathMatching("/", []string{"^/api"}))
	assert.True(t, isPathMatching("/api/test", []string{"^/api"}))
	assert.False(t, isPathMatching("/api/test", []string{"^/api$"}))
	assert.True(t, isPathMatching("/api", []string{"^/api$"}))
	assert.True(t, isPathMatching("/api/test", []string{"^/api/?.*"}))
	assert.True(t, isPathMatching("/apitest", []string{"^/api/?.*"}))
	assert.True(t, isPathMatching("/api/test", []string{"^/api/.*"}))
	assert.True(t, isPathMatching("/api/", []string{"^/api/.*"}))
	assert.False(t, isPathMatching("/api", []string{"^/api/.*"}))

	assert.False(t, isPathMatching("/api", []string{"xyz", "^/api/.*"}))
}
