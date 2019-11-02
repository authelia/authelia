package suites

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

// WebDriverSession binding a selenium service and a webdriver.
type WebDriverSession struct {
	service   *selenium.Service
	WebDriver selenium.WebDriver
}

// StartWebDriver create a selenium session
func StartWebDriver() (*WebDriverSession, error) {
	port := 4444
	service, err := selenium.NewChromeDriverService("/usr/bin/chromedriver", port)

	if err != nil {
		return nil, err
	}

	chromeCaps := chrome.Capabilities{
		Path: "/usr/bin/google-chrome-stable",
	}

	if os.Getenv("HEADLESS") != "" {
		chromeCaps.Args = append(chromeCaps.Args, "--headless")
	}

	caps := selenium.Capabilities{}
	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		service.Stop()
		panic(err)
	}

	return &WebDriverSession{
		service:   service,
		WebDriver: wd,
	}, nil
}

// Stop stop the selenium session
func (wds *WebDriverSession) Stop() error {
	err := wds.WebDriver.Quit()

	if err != nil {
		return err
	}

	return wds.service.Stop()
}

// WithWebdriver run some actions against a webdriver
func WithWebdriver(fn func(webdriver selenium.WebDriver) error) error {
	wds, err := StartWebDriver()

	if err != nil {
		return err
	}

	defer wds.Stop()

	return fn(wds.WebDriver)
}

func waitElementLocated(ctx context.Context, s *SeleniumSuite, by, value string) selenium.WebElement {
	var el selenium.WebElement
	err := s.Wait(ctx, func(driver selenium.WebDriver) (bool, error) {
		var err error
		el, err = driver.FindElement(by, value)

		if err != nil {
			if strings.Contains(err.Error(), "no such element") {
				return false, nil
			}
			return false, err
		}

		return el != nil, nil
	})

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), el, "Element has not been located")
	return el
}

// WaitElementLocatedByID wait an element is located by id
func WaitElementLocatedByID(ctx context.Context, s *SeleniumSuite, id string) selenium.WebElement {
	return waitElementLocated(ctx, s, selenium.ByID, id)
}

// WaitElementLocatedByTagName wait an element is located by tag name
func WaitElementLocatedByTagName(ctx context.Context, s *SeleniumSuite, tagName string) selenium.WebElement {
	return waitElementLocated(ctx, s, selenium.ByTagName, tagName)
}

// WaitElementLocatedByClassName wait an element is located by class name
func WaitElementLocatedByClassName(ctx context.Context, s *SeleniumSuite, className string) selenium.WebElement {
	return waitElementLocated(ctx, s, selenium.ByClassName, className)
}

// WaitElementTextContains wait the text of an element contains a pattern
func WaitElementTextContains(ctx context.Context, s *SeleniumSuite, element selenium.WebElement, pattern string) {
	assert.NotNil(s.T(), element)

	s.Wait(ctx, func(driver selenium.WebDriver) (bool, error) {
		text, err := element.Text()

		if err != nil {
			return false, err
		}

		return strings.Contains(text, pattern), nil
	})
}
