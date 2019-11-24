package suites

import (
	"context"
	"fmt"
	"testing"
)

func (wds *WebDriverSession) verifyIsHome(ctx context.Context, t *testing.T) {
	wds.verifyURLIs(ctx, t, fmt.Sprintf("%s/", HomeBaseURL))
}
