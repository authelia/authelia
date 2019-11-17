package suites

import (
	"testing"
)

type DuoPushSuite struct {
	*SeleniumSuite
}

func NewDuoPushSuite() *DuoPushSuite {
	return &DuoPushSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestDuoPushSuite(t *testing.T) {
	RunTypescriptSuite(t, duoPushSuiteName)
}
