package suites

import (
	"regexp"
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) verifyURLIs(t *testing.T, page *rod.Page, url string) {
	currentURL := page.MustInfo().URL
	require.Equal(t, url, currentURL, "they should be equal")
}

func (rs *RodSession) verifyURLIsRegexp(t *testing.T, page *rod.Page, rx *regexp.Regexp) {
	currentURL := page.MustInfo().URL

	require.Regexp(t, rx, currentURL, "url should match the expression")
}
