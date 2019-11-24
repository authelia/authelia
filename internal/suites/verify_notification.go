package suites

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (wds *WebDriverSession) verifyNotificationDisplayed(ctx context.Context, t *testing.T, message string) {
	el := wds.WaitElementLocatedByClassName(ctx, t, "notification")
	assert.NotNil(t, el)
	wds.WaitElementTextContains(ctx, t, el, message)
}
