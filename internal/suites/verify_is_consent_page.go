package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifyIsOpenIDConsentDecisionStage(t *testing.T, page *rod.Page) {
	rs.WaitElementLocatedByID(t, page, "openid-consent-decision-stage")
}
