package suites

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func (wds *WebDriverSession) doInitiatePasswordReset(ctx context.Context, t *testing.T, username string) {
	err := wds.WaitElementLocatedByID(ctx, t, "reset-password-button").Click()
	require.NoError(t, err)
	// Fill in username
	err = wds.WaitElementLocatedByID(ctx, t, "username-textfield").SendKeys(username)
	require.NoError(t, err)
	// And click on the reset button
	err = wds.WaitElementLocatedByID(ctx, t, "reset-button").Click()
	require.NoError(t, err)
}

func (wds *WebDriverSession) doCompletePasswordReset(ctx context.Context, t *testing.T, newPassword1, newPassword2 string) {
	link := doGetLinkFromLastMail(t)
	wds.doVisit(t, link)

	err := wds.WaitElementLocatedByID(ctx, t, "password1-textfield").SendKeys(newPassword1)
	require.NoError(t, err)
	err = wds.WaitElementLocatedByID(ctx, t, "password2-textfield").SendKeys(newPassword2)
	require.NoError(t, err)
	err = wds.WaitElementLocatedByID(ctx, t, "reset-button").Click()
	require.NoError(t, err)
}

func (wds *WebDriverSession) doSuccessfullyCompletePasswordReset(ctx context.Context, t *testing.T, newPassword1, newPassword2 string) {
	wds.doCompletePasswordReset(ctx, t, newPassword1, newPassword2)
	wds.verifyIsFirstFactorPage(ctx, t)
}

func (wds *WebDriverSession) doUnsuccessfulPasswordReset(ctx context.Context, t *testing.T, newPassword1, newPassword2 string) {
	wds.doCompletePasswordReset(ctx, t, newPassword1, newPassword2)
}

func (wds *WebDriverSession) doResetPassword(ctx context.Context, t *testing.T, username, newPassword1, newPassword2 string, unsuccessful bool) {
	wds.doInitiatePasswordReset(ctx, t, username)
	// then wait for the "email sent notification"
	wds.verifyMailNotificationDisplayed(ctx, t)

	if unsuccessful {
		wds.doUnsuccessfulPasswordReset(ctx, t, newPassword1, newPassword2)
	} else {
		wds.doSuccessfullyCompletePasswordReset(ctx, t, newPassword1, newPassword2)
	}
}
