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

func (rs *RodSession) doChangeDevice(t *testing.T, page *rod.Page, deviceID string) {
	err := rs.WaitElementLocatedByCSSSelector(t, page, "selection-link").Click("left")
	require.NoError(t, err)
	rs.doSelectDevice(t, page, deviceID)
}

func (rs *RodSession) doSelectDevice(t *testing.T, page *rod.Page, deviceID string) {
	rs.WaitElementLocatedByCSSSelector(t, page, "device-selection")
	err := rs.WaitElementLocatedByCSSSelector(t, page, fmt.Sprintf("device-%s", deviceID)).Click("left")
	require.NoError(t, err)
}

func (rs *RodSession) doClickButton(t *testing.T, page *rod.Page, buttonID string) {
	err := rs.WaitElementLocatedByCSSSelector(t, page, buttonID).Click("left")
	require.NoError(t, err)
}
