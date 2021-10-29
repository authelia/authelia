package suites

import (
	"fmt"
	"testing"

	"github.com/go-rod/rod"
)

func (rs *RodSession) doChangeMethod(t *testing.T, page *rod.Page, method string) {
	rs.WaitElementLocatedByID(t, page, "methods-button").MustClick()
	rs.WaitElementLocatedByID(t, page, "methods-dialog")
	rs.WaitElementLocatedByID(t, page, fmt.Sprintf("%s-option", method)).MustClick()
}

func (rs *RodSession) doChangeDevice(t *testing.T, page *rod.Page, deviceID string) {
	rs.WaitElementLocatedByID(t, page, "selection-link").MustClick()
	rs.doSelectDevice(t, page, deviceID)
}

func (rs *RodSession) doSelectDevice(t *testing.T, page *rod.Page, deviceID string) {
	rs.WaitElementLocatedByID(t, page, "device-selection")
	rs.WaitElementLocatedByID(t, page, fmt.Sprintf("device-%s", deviceID)).MustClick()
}

func (rs *RodSession) doClickButton(t *testing.T, page *rod.Page, buttonID string) {
	rs.WaitElementLocatedByID(t, page, buttonID).MustClick()
}
