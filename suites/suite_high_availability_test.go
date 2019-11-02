package suites

import (
	"testing"
)

type HighAvailabilitySuite struct {
	*SeleniumSuite
}

func NewHighAvailabilitySuite() *HighAvailabilitySuite {
	return &HighAvailabilitySuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestHighAvailabilitySuite(t *testing.T) {
	RunTypescriptSuite(t, highAvailabilitySuiteName)

	TestRunOneFactor(t)
}
