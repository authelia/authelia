package suites

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (wds *WebDriverSession) doRegisterTOTP(ctx context.Context, t *testing.T) string {
	err := wds.WaitElementLocatedByID(ctx, t, "register-link").Click()
	require.NoError(t, err)
	wds.verifyMailNotificationDisplayed(ctx, t)
	link := doGetLinkFromLastMail(t)
	wds.doVisit(t, link)
	secretURL, err := wds.WaitElementLocatedByID(ctx, t, "secret-url").GetAttribute("value")
	assert.NoError(t, err)

	secret := secretURL[strings.LastIndex(secretURL, "=")+1:]
	assert.NotEqual(t, "", secret)
	assert.NotNil(t, secret)

	return secret
}

func (wds *WebDriverSession) doEnterOTP(ctx context.Context, t *testing.T, code string) {
	inputs := wds.WaitElementsLocatedByCSSSelector(ctx, t, "#otp-input input")

	for i := 0; i < 6; i++ {
		err := inputs[i].SendKeys(string(code[i]))
		require.NoError(t, err)
	}
}

func (wds *WebDriverSession) doValidateTOTP(ctx context.Context, t *testing.T, secret string) {
	code, err := totp.GenerateCode(secret, time.Now())
	assert.NoError(t, err)
	wds.doEnterOTP(ctx, t, code)
}
