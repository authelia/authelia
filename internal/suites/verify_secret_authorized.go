package suites

import (
	"context"
	"testing"
)

func (wds *WebDriverSession) verifySecretAuthorized(ctx context.Context, t *testing.T) {
	wds.verifyBodyContains(ctx, t, "This is a very important secret!")
}
