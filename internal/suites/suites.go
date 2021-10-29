package suites

import (
	"github.com/go-rod/rod"
)

// RodSuite is a go-rod suite.
type RodSuite struct {
	*RodSession
	*rod.Page
}

// CommandSuite is a command line interface suite.
type CommandSuite struct {
	testArg     string
	coverageArg string

	*DockerEnvironment
}
