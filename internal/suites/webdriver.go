package suites

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

// WebDriverSession binding a selenium service and a webdriver.
type WebDriverSession struct {
	service   *selenium.Service
	WebDriver selenium.WebDriver
}

// StartWebDriverWithProxy create a selenium session.
func StartWebDriverWithProxy(proxy string, port int) (*WebDriverSession, error) {
	driverPath := os.Getenv("CHROMEDRIVER_PATH")
	if driverPath == "" {
		driverPath = "/usr/bin/chromedriver"
	}

	service, err := selenium.NewChromeDriverService(driverPath, port)

	if err != nil {
		return nil, err
	}

	browserPath := os.Getenv("BROWSER_PATH")
	if browserPath == "" {
		browserPath = "/usr/bin/chromium-browser"
	}

	chromeCaps := chrome.Capabilities{
		Path: browserPath,
	}

	chromeCaps.Args = append(chromeCaps.Args, "--ignore-certificate-errors")

	if os.Getenv("HEADLESS") != "" {
		chromeCaps.Args = append(chromeCaps.Args, "--headless")
		chromeCaps.Args = append(chromeCaps.Args, "--no-sandbox")
	}

	if proxy != "" {
		chromeCaps.Args = append(chromeCaps.Args, fmt.Sprintf("--proxy-server=%s", proxy))
	}

	caps := selenium.Capabilities{}
	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		_ = service.Stop()

		log.Fatal(err)
	}

	return &WebDriverSession{
		service:   service,
		WebDriver: wd,
	}, nil
}

// StartWebDriver create a selenium session.
func StartWebDriver() (*WebDriverSession, error) {
	return StartWebDriverWithProxy("", GetWebDriverPort())
}

// Stop stop the selenium session.
func (wds *WebDriverSession) Stop() error {
	var coverage map[string]interface{}

	coverageDir := "../../web/.nyc_output"
	time := time.Now()

	resp, err := wds.WebDriver.ExecuteScriptRaw("return JSON.stringify(window.__coverage__)", nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, &coverage)
	if err != nil {
		return err
	}

	coverageData := fmt.Sprintf("%s", coverage["value"])

	_ = os.MkdirAll(coverageDir, 0775)

	err = ioutil.WriteFile(fmt.Sprintf("%s/coverage-%d.json", coverageDir, time.Unix()), []byte(coverageData), 0664) //nolint:gosec
	if err != nil {
		return err
	}

	err = wds.WebDriver.Quit()
	if err != nil {
		return err
	}

	return wds.service.Stop()
}

// WithWebdriver run some actions against a webdriver.
func WithWebdriver(fn func(webdriver selenium.WebDriver) error) error {
	wds, err := StartWebDriver()

	if err != nil {
		return err
	}

	defer wds.Stop() //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	return fn(wds.WebDriver)
}

// Wait wait until condition holds true.
func (wds *WebDriverSession) Wait(ctx context.Context, condition selenium.Condition) error {
	done := make(chan error, 1)

	go func() {
		done <- wds.WebDriver.Wait(condition)
	}()

	select {
	case <-ctx.Done():
		return errors.New("waiting timeout reached")
	case err := <-done:
		return err
	}
}

func (wds *WebDriverSession) waitElementLocated(ctx context.Context, t *testing.T, by, value string) selenium.WebElement {
	var el selenium.WebElement

	err := wds.Wait(ctx, func(driver selenium.WebDriver) (bool, error) {
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

	require.NoError(t, err)
	require.NotNil(t, el)

	return el
}

func (wds *WebDriverSession) waitElementsLocated(ctx context.Context, t *testing.T, by, value string) []selenium.WebElement {
	var el []selenium.WebElement

	err := wds.Wait(ctx, func(driver selenium.WebDriver) (bool, error) {
		var err error
		el, err = driver.FindElements(by, value)

		if err != nil {
			if strings.Contains(err.Error(), "no such element") {
				return false, nil
			}
			return false, err
		}

		return el != nil, nil
	})

	require.NoError(t, err)
	require.NotNil(t, el)

	return el
}

// WaitElementLocatedByID wait an element is located by id.
func (wds *WebDriverSession) WaitElementLocatedByID(ctx context.Context, t *testing.T, id string) selenium.WebElement {
	return wds.waitElementLocated(ctx, t, selenium.ByID, id)
}

// WaitElementLocatedByTagName wait an element is located by tag name.
func (wds *WebDriverSession) WaitElementLocatedByTagName(ctx context.Context, t *testing.T, tagName string) selenium.WebElement {
	return wds.waitElementLocated(ctx, t, selenium.ByTagName, tagName)
}

// WaitElementLocatedByClassName wait an element is located by class name.
func (wds *WebDriverSession) WaitElementLocatedByClassName(ctx context.Context, t *testing.T, className string) selenium.WebElement {
	return wds.waitElementLocated(ctx, t, selenium.ByClassName, className)
}

// WaitElementLocatedByLinkText wait an element is located by link text.
func (wds *WebDriverSession) WaitElementLocatedByLinkText(ctx context.Context, t *testing.T, linkText string) selenium.WebElement {
	return wds.waitElementLocated(ctx, t, selenium.ByLinkText, linkText)
}

// WaitElementLocatedByCSSSelector wait an element is located by class name.
func (wds *WebDriverSession) WaitElementLocatedByCSSSelector(ctx context.Context, t *testing.T, cssSelector string) selenium.WebElement {
	return wds.waitElementLocated(ctx, t, selenium.ByCSSSelector, cssSelector)
}

// WaitElementsLocatedByCSSSelector wait an element is located by CSS selector.
func (wds *WebDriverSession) WaitElementsLocatedByCSSSelector(ctx context.Context, t *testing.T, cssSelector string) []selenium.WebElement {
	return wds.waitElementsLocated(ctx, t, selenium.ByCSSSelector, cssSelector)
}

// WaitElementTextContains wait the text of an element contains a pattern.
func (wds *WebDriverSession) WaitElementTextContains(ctx context.Context, t *testing.T, element selenium.WebElement, pattern string) {
	err := wds.Wait(ctx, func(driver selenium.WebDriver) (bool, error) {
		text, err := element.Text()

		if err != nil {
			return false, err
		}

		return strings.Contains(text, pattern), nil
	})
	require.NoError(t, err)
}
