package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestTraefik2Suite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTraefikSuite(traefik2SuiteName))
}
