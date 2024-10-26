package suites

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type KubernetesSuite struct {
	*RodSuite
}

func NewKubernetesSuite() *KubernetesSuite {
	return &KubernetesSuite{
		RodSuite: NewRodSuite(kubernetesSuiteName),
	}
}

func (s *KubernetesSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *KubernetesSuite) TestTwoFactorTOTPScenario() {
	suite.Run(s.T(), NewTwoFactorTOTPScenario())
}

func (s *KubernetesSuite) TestRedirectionURLScenario() {
	suite.Run(s.T(), NewRedirectionURLScenario())
}

func TestKubernetesSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewKubernetesSuite())
}
