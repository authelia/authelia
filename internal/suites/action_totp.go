package suites

import (
	"context"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
)

func (wds *WebDriverSession) doRegisterTOTP(ctx context.Context, t *testing.T) string {
	wds.WaitElementLocatedByID(ctx, t, "register-link").Click()
	wds.verifyMailNotificationDisplayed(ctx, t)
	link := doGetLinkFromLastMail(t)
	wds.doVisit(t, link)
	secret, err := wds.WaitElementLocatedByID(ctx, t, "base32-secret").GetAttribute("value")
	assert.NoError(t, err)
	assert.NotEqual(t, "", secret)
	assert.NotNil(t, secret)
	return secret
}

func (wds *WebDriverSession) doEnterOTP(ctx context.Context, t *testing.T, code string) {
	inputs := wds.WaitElementsLocatedByCSSSelector(ctx, t, "#otp-input input")

	for i := 0; i < 6; i++ {
		inputs[i].SendKeys(string(code[i]))
	}
}

func (wds *WebDriverSession) doValidateTOTP(ctx context.Context, t *testing.T, secret string) {
	code, err := totp.GenerateCode(secret, time.Now())
	assert.NoError(t, err)
	wds.doEnterOTP(ctx, t, code)
}
