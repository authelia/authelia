package suites

import (
	"testing"
)

type BypassAllSuite struct {
	*SeleniumSuite
}

func NewBypassAllSuite() *BypassAllSuite {
	return &BypassAllSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestBypassAllSuite(t *testing.T) {
	RunTypescriptSuite(t, bypassAllSuiteName)
}
