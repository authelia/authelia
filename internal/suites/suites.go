package suites

import (
	"github.com/go-rod/rod"
	"github.com/stretchr/testify/suite"
)

// NewRodSuite returns a new *RodSuite with the given name.
func NewRodSuite(name string) *RodSuite {
	return &RodSuite{
		BaseSuite: &BaseSuite{
			Name: name,
		},
		RodSuiteCredentialsProvider: NewRodSuiteCredentials(),
	}
}

// RodSuite is a go-rod suite.
type RodSuite struct {
	*BaseSuite

	*RodSession
	*rod.Page

	RodSuiteCredentialsProvider
}

// MustClose closes the page when the suite has one. It shadows the method promoted from the embedded
// page so that a teardown following a setup which failed before the tab was created closes nothing
// instead of dereferencing a page that was never assigned.
func (s *RodSuite) MustClose() {
	if s.Page == nil {
		return
	}

	s.Page.MustClose()
}

// BaseSuite is the base suite which every suite embeds.
type BaseSuite struct {
	suite.Suite

	Name string
}

// CommandSuite is a command line interface suite.
type CommandSuite struct {
	*BaseSuite

	*DockerEnvironment
}
