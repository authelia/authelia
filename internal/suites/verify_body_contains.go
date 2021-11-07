package suites

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) verifyBodyContains(t *testing.T, page *rod.Page, pattern string) {
	body, err := page.Element("body")
	assert.NoError(t, err)
	assert.NotNil(t, body)

	text, err := body.Text()
	assert.NoError(t, err)
	assert.NotNil(t, text)

	if strings.Contains(text, pattern) {
		err = nil
	} else {
		err = fmt.Errorf("body does not contain pattern: %s", pattern)
	}

	require.NoError(t, err)
}
