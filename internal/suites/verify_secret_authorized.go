package suites

import "context"

func verifySecretAuthorized(ctx context.Context, s *SeleniumSuite) {
	verifyBodyContains(ctx, s, "This is a very important secret!")
}
