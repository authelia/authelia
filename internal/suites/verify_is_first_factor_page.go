package suites

import (
	"context"
	"testing"
)

func (wds *WebDriverSession) verifyIsFirstFactorPage(ctx context.Context, t *testing.T) {
	wds.WaitElementLocatedByID(ctx, t, "first-factor-stage")
}
