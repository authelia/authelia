package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifyIsOIDC(t *testing.T, page *rod.Page, pattern, url string) {
	page.MustElementR("body", pattern)
	rs.verifyURLIs(t, page, url)
}
