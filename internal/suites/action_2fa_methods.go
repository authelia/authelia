package suites

import (
	"fmt"
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) doChangeMethod(t *testing.T, page *rod.Page, method string) {
	rs.ClickElementLocatedByID(t, page, "methods-button")
	rs.WaitElementLocatedByID(t, page, "methods-dialog")
	rs.ClickElementLocatedByID(t, page, fmt.Sprintf("%s-option", method))
}

func (rs *RodSession) doChangeDevice(t *testing.T, page *rod.Page, deviceID string) {
	rs.ClickElementLocatedByID(t, page, "selection-link")
	rs.doSelectDevice(t, page, deviceID)
}

func (rs *RodSession) doSelectDevice(t *testing.T, page *rod.Page, deviceID string) {
	rs.WaitElementLocatedByID(t, page, "device-selection")
	rs.ClickElementLocatedByID(t, page, fmt.Sprintf("device-%s", deviceID))
}

func (rs *RodSession) doClickButton(t *testing.T, page *rod.Page, buttonID string) {
	rs.ClickElementLocatedByID(t, page, buttonID)
}
