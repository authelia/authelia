package suites

import (
	"github.com/go-rod/rod"
	"github.com/stretchr/testify/suite"
)

// RodSuite is a go-rod suite.
type RodSuite struct {
	suite.Suite

	*RodSession
	*rod.Page
}

// CommandSuite is a command line interface suite.
type CommandSuite struct {
	suite.Suite

	testArg     string //nolint:structcheck // TODO: Remove when bug fixed: https://github.com/golangci/golangci-lint/issues/537.
	coverageArg string //nolint:structcheck // TODO: Remove when bug fixed: https://github.com/golangci/golangci-lint/issues/537.

	*DockerEnvironment
}
