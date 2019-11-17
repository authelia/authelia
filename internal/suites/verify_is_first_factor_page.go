package suites

import "context"

func verifyIsFirstFactorPage(ctx context.Context, s *SeleniumSuite) {
	WaitElementLocatedByClassName(ctx, s, "first-factor-step")
}
