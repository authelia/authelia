package suites

import "context"

func doRegisterThenLogout(ctx context.Context, s *SeleniumSuite, username, password string) string {
	secret := doLoginAndRegisterTOTP(ctx, s, username, password, false)
	doLogout(ctx, s)
	return secret
}
