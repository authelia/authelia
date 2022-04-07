package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifySecretAuthorized(t *testing.T, page *rod.Page) {
	rs.WaitElementLocatedByID(t, page, "secret")
}
