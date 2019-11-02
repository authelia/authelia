package suites

import "context"

func verifyIsSecondFactorPage(ctx context.Context, s *SeleniumSuite) {
	WaitElementLocatedByClassName(ctx, s, "second-factor-step")
}
