package suites

import (
	"context"
	"testing"
)

func (wds *WebDriverSession) verifyMailNotificationDisplayed(ctx context.Context, t *testing.T) {
	wds.verifyNotificationDisplayed(ctx, t, "An email has been sent to your address to complete the process.")
}
