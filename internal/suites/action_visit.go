package suites

import (
	"context"
	"fmt"
	"net/url"

	"github.com/stretchr/testify/assert"
)

func doVisit(s *SeleniumSuite, url string) {
	err := s.WebDriver().Get(url)
	assert.NoError(s.T(), err)
}

func doVisitAndVerifyURLIs(ctx context.Context, s *SeleniumSuite, url string) {
	doVisit(s, url)
	verifyURLIs(ctx, s, url)
}

func doVisitLoginPage(ctx context.Context, s *SeleniumSuite, targetURL string) {
	suffix := ""
	if targetURL != "" {
		suffix = fmt.Sprintf("?rd=%s", url.QueryEscape(targetURL))
	}
	doVisitAndVerifyURLIs(ctx, s, fmt.Sprintf("%s%s", LoginBaseURL, suffix))
}
