package suites

import "context"

func verifyNotificationDisplayed(ctx context.Context, s *SeleniumSuite, message string) {
	txt, err := WaitElementLocatedByClassName(ctx, s, "notification").Text()
	s.Assert().NoError(err)
	s.Assert().Equal(message, txt)
}
