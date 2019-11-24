package suites

import (
	"context"
	"fmt"
	"testing"
)

func (wds *WebDriverSession) doChangeMethod(ctx context.Context, t *testing.T, method string) {
	wds.WaitElementLocatedByID(ctx, t, "methods-button").Click()
	wds.WaitElementLocatedByID(ctx, t, fmt.Sprintf("%s-option", method)).Click()
}
