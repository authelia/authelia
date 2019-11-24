package suites

import (
	"context"
	"testing"
)

func (wds *WebDriverSession) doRegisterThenLogout(ctx context.Context, t *testing.T, username, password string) string {
	secret := wds.doLoginAndRegisterTOTP(ctx, t, username, password, false)
	wds.doLogout(ctx, t)
	return secret
}
