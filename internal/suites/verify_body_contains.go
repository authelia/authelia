package suites

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/assert"
)

func (rs *RodSession) verifyBodyContains(t *testing.T, page *rod.Page, pattern string) {
	body, err := page.ElementR("body", pattern)
	assert.NoError(t, err)
	assert.NotNil(t, body)
}
