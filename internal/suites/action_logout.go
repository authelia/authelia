package suites

import (
	"context"
	"fmt"
	"testing"
)

func (wds *WebDriverSession) doLogout(ctx context.Context, t *testing.T) {
	if PathPrefix != "" {
		wds.doVisit(t, fmt.Sprintf("%s%s%s", LoginBaseURL, PathPrefix, "/logout"))
	} else {
		wds.doVisit(t, fmt.Sprintf("%s%s", LoginBaseURL, "/logout"))
	}

	wds.verifyIsFirstFactorPage(ctx, t)
}
