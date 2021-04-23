package suites

import (
	"context"
	"testing"
)

func (wds *WebDriverSession) verifyIsConsentPage(ctx context.Context, t *testing.T) {
	wds.WaitElementLocatedByID(ctx, t, "consent-stage")
}
