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

func (wds *WebDriverSession) doLogoutWithRedirect(ctx context.Context, t *testing.T, targetURL string, firstFactor bool) {
	wds.doVisit(t, fmt.Sprintf("%s%s%s", GetLoginBaseURL(), "/logout?rd=", url.QueryEscape(targetURL)))

	fmt.Printf("DEBUG: Logout Target URL is: %s, firstFactor: %t\n", targetURL, firstFactor)

	if firstFactor {
		wds.verifyIsFirstFactorPage(ctx, t)

		return
	}

	wds.verifyURLIs(ctx, t, targetURL)
}
