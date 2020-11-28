package suites

import (
	"context"
	"testing"
)

func (wds *WebDriverSession) doInitiatePasswordReset(ctx context.Context, t *testing.T, username string) {
	wds.WaitElementLocatedByID(ctx, t, "reset-password-button").Click() //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	// Fill in username
	wds.WaitElementLocatedByID(ctx, t, "username-textfield").SendKeys(username) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	// And click on the reset button
	wds.WaitElementLocatedByID(ctx, t, "reset-button").Click() //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
}

func (wds *WebDriverSession) doCompletePasswordReset(ctx context.Context, t *testing.T, newPassword1, newPassword2 string) {
	link := doGetLinkFromLastMail(t)
	wds.doVisit(t, link)

	wds.WaitElementLocatedByID(ctx, t, "password1-textfield").SendKeys(newPassword1) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	wds.WaitElementLocatedByID(ctx, t, "password2-textfield").SendKeys(newPassword2) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
	wds.WaitElementLocatedByID(ctx, t, "reset-button").Click()                       //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
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
