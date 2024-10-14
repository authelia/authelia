package suites

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) verifyIsResetPasswordPage(t *testing.T, page *rod.Page) {
	require.NoError(t, page.WaitStable(time.Millisecond*50))

	rs.WaitElementLocatedByID(t, page, "reset-password-step1-stage")
}
