package suites

import (
	"context"
	"fmt"
	"net/url"
	"testing"
)

func (wds *WebDriverSession) doLogout(ctx context.Context, t *testing.T) {
	wds.doVisit(t, fmt.Sprintf("%s%s", GetLoginBaseURL(), "/logout"))
	wds.verifyIsFirstFactorPage(ctx, t)
}

func (wds *WebDriverSession) doLogoutWithRedirect(ctx context.Context, t *testing.T, redirectURL string, firstFactor bool) {
	// TODO: REMOVE.
	fmt.Printf("DEBUG: Attempting Logout With redirect to: %s\n", redirectURL)

	wds.doVisit(t, fmt.Sprintf("%s%s%s", GetLoginBaseURL(), "/logout?rd=", url.QueryEscape(redirectURL)))

	if firstFactor {
		// TODO: REMOVE.
		fmt.Printf("DEBUG: Attempting Logout Expected Page is First Factor\n")

		wds.verifyIsFirstFactorPage(ctx, t)
		return
	}

	// TODO: REMOVE.
	fmt.Printf("DEBUG: Attempting Logout Expected Page is %s\n", redirectURL)

	wds.verifyURLIs(ctx, t, redirectURL)
}
