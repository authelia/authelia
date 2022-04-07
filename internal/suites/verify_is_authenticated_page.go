package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifyIsAuthenticatedPage(t *testing.T, page *rod.Page) {
	rs.WaitElementLocatedByID(t, page, "authenticated-stage")
}
