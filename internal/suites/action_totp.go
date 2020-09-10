package suites

import (
	"context"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
)

func (wds *WebDriverSession) doRegisterTOTP(ctx context.Context, t *testing.T) string {
	wds.WaitElementLocatedByID(ctx, t, "register-link").Click() //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
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
		inputs[i].SendKeys(string(code[i])) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	}
}

func (wds *WebDriverSession) doValidateTOTP(ctx context.Context, t *testing.T, secret string) {
	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}

	code, err := totp.GenerateCodeCustom(secret, time.Now(), opts)
	assert.NoError(t, err)
	wds.doEnterOTP(ctx, t, code)
}
