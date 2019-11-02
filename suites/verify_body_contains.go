package suites

import "context"

func verifyBodyContains(ctx context.Context, s *SeleniumSuite, pattern string) {
	bodyElement := WaitElementLocatedByTagName(ctx, s, "body")
	WaitElementTextContains(ctx, s, bodyElement, pattern)
}
