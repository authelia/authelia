package suites

import (
	"github.com/stretchr/testify/suite"
	"github.com/tebeka/selenium"
)

// SeleniumSuite is a selenium suite
type SeleniumSuite struct {
	suite.Suite

	*WebDriverSession
}

// WebDriver return the webdriver of the suite
func (s *SeleniumSuite) WebDriver() selenium.WebDriver {
	return s.WebDriverSession.WebDriver
}
