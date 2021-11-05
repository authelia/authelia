package suites

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) verifyURLIs(t *testing.T, page *rod.Page, url string) {
	currentURL := page.MustInfo().URL
	require.Equal(t, url, currentURL, "they should be equal")
}
