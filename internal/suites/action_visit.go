package suites

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (wds *WebDriverSession) doVisit(t *testing.T, url string) {
	err := wds.WebDriver.Get(url)
	assert.NoError(t, err)
}

func (wds *WebDriverSession) doVisitAndVerifyOneFactorStep(ctx context.Context, t *testing.T, url string) {
	wds.doVisit(t, url)
	wds.verifyIsFirstFactorPage(ctx, t)
}

func (wds *WebDriverSession) doVisitLoginPage(ctx context.Context, t *testing.T, targetURL string) {
	suffix := ""
	if targetURL != "" {
		suffix = fmt.Sprintf("?rd=%s", targetURL)
	}
	wds.doVisitAndVerifyOneFactorStep(ctx, t, fmt.Sprintf("%s/%s", LoginBaseURL, suffix))
}
