package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifyIsSecondFactorPage(t *testing.T, page *rod.Page) {
	rs.WaitElementLocatedByID(t, page, "second-factor-stage")
}
