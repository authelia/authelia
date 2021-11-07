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
