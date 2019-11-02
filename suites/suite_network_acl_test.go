package suites

import (
	"testing"
)

type NetworkACLSuite struct {
	*SeleniumSuite
}

func NewNetworkACLSuite() *NetworkACLSuite {
	return &NetworkACLSuite{SeleniumSuite: new(SeleniumSuite)}
}

func TestNetworkACLSuite(t *testing.T) {
	RunTypescriptSuite(t, networkACLSuiteName)
}
