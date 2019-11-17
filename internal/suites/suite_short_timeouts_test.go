package suites

import (
	"testing"
)

type ShortTimeoutsSuite struct {
	*SeleniumSuite
}

func NewShortTimeoutsSuite() *ShortTimeoutsSuite {
	return &ShortTimeoutsSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestShortTimeoutsSuite(t *testing.T) {
	RunTypescriptSuite(t, shortTimeoutsSuiteName)
}
