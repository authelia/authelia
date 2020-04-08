package suites

import (
	"context"
	"testing"
)

func (wds *WebDriverSession) verifyIsAuthenticatedPage(ctx context.Context, t *testing.T) {
	wds.WaitElementLocatedByID(ctx, t, "authenticated-stage")
}
