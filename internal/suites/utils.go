package suites

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func fixCoveragePath(path string, file os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if file.IsDir() {
		return nil
	}

	coverage, err := filepath.Match("*.json", file.Name())

	if err != nil {
		return err
	}

	if coverage {
		read, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		wd, _ := os.Getwd()
		ciPath := strings.TrimSuffix(wd, "internal/suites")
		content := strings.ReplaceAll(string(read), "/node/src/app/", ciPath+"web/")

		err = ioutil.WriteFile(path, []byte(content), 0)
		if err != nil {
			return err
		}
	}

	return nil
}
