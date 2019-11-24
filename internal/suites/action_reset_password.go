package suites

import (
	"context"
	"testing"
)

func (wds *WebDriverSession) doInitiatePasswordReset(ctx context.Context, t *testing.T, username string) {
	wds.WaitElementLocatedByID(ctx, t, "reset-password-button").Click()
	// Fill in username
	wds.WaitElementLocatedByID(ctx, t, "username-textfield").SendKeys(username)
	// And click on the reset button
	wds.WaitElementLocatedByID(ctx, t, "reset-button").Click()
}

func (wds *WebDriverSession) doCompletePasswordReset(ctx context.Context, t *testing.T, newPassword1, newPassword2 string) {
	link := doGetLinkFromLastMail(t)
	wds.doVisit(t, link)

	wds.WaitElementLocatedByID(ctx, t, "password1-textfield").SendKeys(newPassword1)
	wds.WaitElementLocatedByID(ctx, t, "password2-textfield").SendKeys(newPassword2)
	wds.WaitElementLocatedByID(ctx, t, "reset-button").Click()
}

func (wds *WebDriverSession) doSuccessfullyCompletePasswordReset(ctx context.Context, t *testing.T, newPassword1, newPassword2 string) {
	wds.doCompletePasswordReset(ctx, t, newPassword1, newPassword2)
	wds.verifyIsFirstFactorPage(ctx, t)
}

func (wds *WebDriverSession) doResetPassword(ctx context.Context, t *testing.T, username, newPassword1, newPassword2 string) {
	wds.doInitiatePasswordReset(ctx, t, username)
	// then wait for the "email sent notification"
	wds.verifyMailNotificationDisplayed(ctx, t)
	wds.doSuccessfullyCompletePasswordReset(ctx, t, newPassword1, newPassword2)
}
