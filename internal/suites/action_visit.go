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

func (wds *WebDriverSession) doVisitAndVerifyURLIs(ctx context.Context, t *testing.T, url string) {
	wds.doVisit(t, url)
	wds.verifyURLIs(ctx, t, url)
}

func (wds *WebDriverSession) doVisitLoginPage(ctx context.Context, t *testing.T, targetURL string) {
	suffix := ""
	if targetURL != "" {
		suffix = fmt.Sprintf("?rd=%s", targetURL)
	}
	wds.doVisitAndVerifyURLIs(ctx, t, fmt.Sprintf("%s/%s", LoginBaseURL, suffix))
}
