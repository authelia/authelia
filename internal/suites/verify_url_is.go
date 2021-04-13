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

		fmt.Printf("DEBUG: currentURL: %s, expectedURL: %s\n", currentURL, url)

		return currentURL == url, nil
	})

	require.NoError(t, err)
}
