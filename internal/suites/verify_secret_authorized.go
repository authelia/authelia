package suites

import (
	"context"
	"testing"
)

func (wds *WebDriverSession) verifySecretAuthorized(ctx context.Context, t *testing.T) {
	wds.WaitElementLocatedByID(ctx, t, "secret")
}
