package suites

import (
	"context"
	"fmt"
	"testing"
)

func (wds *WebDriverSession) doLogout(ctx context.Context, t *testing.T) {
	wds.doVisit(t, fmt.Sprintf("%s%s", LoginBaseURL, "/logout"))
	wds.verifyIsFirstFactorPage(ctx, t)
}
