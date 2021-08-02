package suites

import (
	"fmt"
	"testing"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doChangeMethod(t *testing.T, page *rod.Page, method string) {
	err := rs.WaitElementLocatedByCSSSelector(t, page, "methods-button").Click("left")
	require.NoError(t, err)
	rs.WaitElementLocatedByCSSSelector(t, page, "methods-dialog")
	err = rs.WaitElementLocatedByCSSSelector(t, page, fmt.Sprintf("%s-option", method)).Click("left")
	require.NoError(t, err)
}

func (wds *WebDriverSession) doChangeDevice(ctx context.Context, t *testing.T, deviceID string) {
	err := wds.WaitElementLocatedByID(ctx, t, "selection-link").Click()
	require.NoError(t, err)
	wds.doSelectDevice(ctx, t, deviceID)
}

func (wds *WebDriverSession) doSelectDevice(ctx context.Context, t *testing.T, deviceID string) {
	wds.WaitElementLocatedByID(ctx, t, "device-selection")
	err := wds.WaitElementLocatedByID(ctx, t, fmt.Sprintf("device-%s", deviceID)).Click()
	require.NoError(t, err)
}

func (wds *WebDriverSession) doClickButton(ctx context.Context, t *testing.T, backID string) {
	err := wds.WaitElementLocatedByID(ctx, t, backID).Click()
	require.NoError(t, err)
}
