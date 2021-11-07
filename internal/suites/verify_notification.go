package suites

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/assert"
)

func (rs *RodSession) verifyNotificationDisplayed(t *testing.T, page *rod.Page, message string) {
	el, err := page.ElementR(".notification", message)
	assert.NoError(t, err)
	assert.NotNil(t, el)
}
