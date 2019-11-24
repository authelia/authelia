package suites

import (
	"context"
	"testing"
)

func (wds *WebDriverSession) verifyIsSecondFactorPage(ctx context.Context, t *testing.T) {
	wds.WaitElementLocatedByID(ctx, t, "second-factor-stage")
}
