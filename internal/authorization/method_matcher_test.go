package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethodMatcher(t *testing.T) {
	assert.False(t, isMethodMatching("", []string{"GET"}))

	assert.False(t, isMethodMatching("GET", []string{"POST", "OPTIONS"}))

	assert.True(t, isMethodMatching("", []string{}))

	assert.True(t, isMethodMatching("GET", []string{"POST", "GET"}))
}
