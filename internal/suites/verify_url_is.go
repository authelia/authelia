package suites

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

func (wds *WebDriverSession) verifyURLIs(ctx context.Context, t *testing.T, url string) {

	// TODO: REMOVE.
	fmt.Printf("DEBUG: Verifying URL is: %s\n", url)

	err := wds.Wait(ctx, func(driver selenium.WebDriver) (bool, error) {
		currentURL, err := driver.CurrentURL()

		// TODO: REMOVE.
		fmt.Printf("DEBUG: Current URL is: %s\n", currentURL)

		if err != nil {
			// TODO: REMOVE.
			fmt.Printf("DEBUG: Err for Current URL is: %v\n", err)

			return false, err
		}
		return currentURL == url, nil
	})

	require.NoError(t, err)
}
