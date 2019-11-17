package suites

import "context"

func doLogout(ctx context.Context, s *SeleniumSuite) {
	doVisit(s, "https://login.example.com:8080/#/logout")
	verifyIsFirstFactorPage(ctx, s)
}
