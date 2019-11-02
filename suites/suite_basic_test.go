package suites

import (
	"testing"
)

type BasicSuite struct {
	*SeleniumSuite
}

func NewBasicSuite() *BasicSuite {
	return &BasicSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestBasicSuite(t *testing.T) {
	RunTypescriptSuite(t, basicSuiteName)
}
