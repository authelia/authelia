package suites

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/matryer/is"
)

func (rs *RodSession) verifyURLIs(t *testing.T, page *rod.Page, url string) {
	is := is.New(t)
	currentURL := page.MustInfo().URL
	is.Equal(url, currentURL)
}
