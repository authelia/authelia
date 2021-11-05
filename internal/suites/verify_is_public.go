package suites

import (
	"fmt"
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) verifyIsPublic(t *testing.T, page *rod.Page) {
	page.MustElementR("body", "headers")
	rs.verifyURLIs(t, page, fmt.Sprintf("%s/headers", PublicBaseURL))
}
