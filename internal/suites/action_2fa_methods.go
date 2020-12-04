package suites

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func (wds *WebDriverSession) doChangeMethod(ctx context.Context, t *testing.T, method string) {
	err := wds.WaitElementLocatedByID(ctx, t, "methods-button").Click()
	require.NoError(t, err)
	wds.WaitElementLocatedByID(ctx, t, "methods-dialog")
	err = wds.WaitElementLocatedByID(ctx, t, fmt.Sprintf("%s-option", method)).Click()
	require.NoError(t, err)
}
