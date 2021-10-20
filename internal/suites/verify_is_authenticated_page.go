package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifyIsAuthenticatedPage(t *testing.T, page *rod.Page) {
	rs.WaitElementLocatedByCSSSelector(t, page, "authenticated-stage")
}
