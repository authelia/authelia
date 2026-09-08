package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifyIsFirstFactorPage(t *testing.T, page *rod.Page) {
	rs.WaitElementLocatedByID(t, page, "first-factor-stage")
}
