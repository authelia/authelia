package suites

import (
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func verifyURLIs(ctx context.Context, s *SeleniumSuite, url string) {
	err := s.Wait(ctx, func(driver selenium.WebDriver) (bool, error) {
		currentURL, err := driver.CurrentURL()

		if err != nil {
			return false, err
		}

		return currentURL == url, nil
	})

	assert.NoError(s.T(), err)
}
