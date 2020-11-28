package suites

import (
	"os"
	"strconv"
)

// GetLoginBaseURL returns the URL of the login portal and the path prefix if specified.
func GetLoginBaseURL() string {
	if PathPrefix != "" {
		return LoginBaseURL + PathPrefix
	}

	return LoginBaseURL
}

// GetWebDriverPort returns the port to initialize the webdriver with.
func GetWebDriverPort() int {
	driverPort := os.Getenv("CHROMEDRIVER_PORT")
	if driverPort == "" {
		driverPort = defaultChromeDriverPort
	}

	p, _ := strconv.Atoi(driverPort)

	return p
}
