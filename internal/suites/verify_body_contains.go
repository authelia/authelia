package suites

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

func (wds *WebDriverSession) verifyBodyContains(ctx context.Context, t *testing.T, pattern string) {
	err := wds.Wait(ctx, func(wd selenium.WebDriver) (bool, error) {
		bodyElement, err := wds.WebDriver.FindElement(selenium.ByTagName, "body")

		if err != nil {
			return false, err
		}

		if bodyElement == nil {
			return false, nil
		}

		content, err := bodyElement.Text()

		if err != nil {
			return false, err
		}

		return strings.Contains(content, pattern), nil
	})
	require.NoError(t, err)
}
