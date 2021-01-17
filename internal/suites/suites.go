package suites

import (
	"github.com/stretchr/testify/suite"
	"github.com/tebeka/selenium"
)

// SeleniumSuite is a selenium suite.
type SeleniumSuite struct {
	suite.Suite

	*WebDriverSession
}

// CommandSuite is a command line interface suite.
type CommandSuite struct {
	suite.Suite

	testArg     string //nolint:structcheck // TODO: Remove when bug fixed: https://github.com/golangci/golangci-lint/issues/537.
	coverageArg string //nolint:structcheck // TODO: Remove when bug fixed: https://github.com/golangci/golangci-lint/issues/537.

	*DockerEnvironment
}

// WebDriver return the webdriver of the suite.
func (s *SeleniumSuite) WebDriver() selenium.WebDriver {
	return s.WebDriverSession.WebDriver
}
