package suites

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/matryer/is"
)

func (rs *RodSession) verifyNotificationDisplayed(t *testing.T, page *rod.Page, message string) {
	is := is.New(t)
	el := page.MustElementR(".notification", message)
	is.True(el != nil)
}
