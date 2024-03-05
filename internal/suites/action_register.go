package suites

import (
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) doLoginAndRegisterTOTPThenLogout(t *testing.T, page *rod.Page, username, password string) {
	rs.doLoginAndRegisterTOTP(t, page, username, password, false)
	rs.doLogout(t, page)
}
