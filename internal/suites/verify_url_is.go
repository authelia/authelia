package suites

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

func (wds *WebDriverSession) verifyURLIs(ctx context.Context, t *testing.T, url string) {
	err := wds.Wait(ctx, func(driver selenium.WebDriver) (bool, error) {
		currentURL, err := driver.CurrentURL()

		if err != nil {
			return false, err
		}
		return currentURL == url, nil
	})

	if err != nil {
		currentURL, _ := wds.WebDriver.CurrentURL()
		fmt.Printf("expected %s != current %s", url, currentURL)
	}

	require.NoError(t, err)
}
