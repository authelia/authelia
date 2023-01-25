package suites

import (
	"github.com/go-rod/rod"
	"github.com/stretchr/testify/suite"
)

func NewRodSuite(name string) *RodSuite {
	return &RodSuite{
		BaseSuite: &BaseSuite{
			Name: name,
		},
	}
}

// RodSuite is a go-rod suite.
type RodSuite struct {
	*BaseSuite

	*RodSession
	*rod.Page
}

type BaseSuite struct {
	suite.Suite

	Name string
}

// CommandSuite is a command line interface suite.
type CommandSuite struct {
	*BaseSuite

	testArg     string //nolint:structcheck // TODO: Remove when bug fixed: https://github.com/golangci/golangci-lint/issues/537.
	coverageArg string //nolint:structcheck // TODO: Remove when bug fixed: https://github.com/golangci/golangci-lint/issues/537.

	*DockerEnvironment
}
