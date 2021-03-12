package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type NetworkACLSuite struct {
	suite.Suite
}

func NewNetworkACLSuite() *NetworkACLSuite {
	return &NetworkACLSuite{}
}

func (s *NetworkACLSuite) TestShouldAccessSecretUpon2FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wds, err := StartWebDriver()
	s.Require().NoError(err)

	defer func() {
		err = wds.Stop()
		s.Require().NoError(err)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SecureBaseURL)
	wds.doVisit(s.T(), targetURL)
	wds.verifyIsFirstFactorPage(ctx, s.T())

	wds.doRegisterAndLogin2FA(ctx, s.T(), "john", "password", false, targetURL)
	wds.verifySecretAuthorized(ctx, s.T())
}

// from network 192.168.240.201/32.
func (s *NetworkACLSuite) TestShouldAccessSecretUpon1FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wds, err := StartWebDriverWithProxy("http://proxy-client1.example.com:3128", GetWebDriverPort())
	s.Require().NoError(err)

	defer func() {
		err = wds.Stop()
		s.Require().NoError(err)
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SecureBaseURL)
	wds.doVisit(s.T(), targetURL)
	wds.verifyIsFirstFactorPage(ctx, s.T())

	wds.doLoginOneFactor(ctx, s.T(), "john", "password",
		false, fmt.Sprintf("%s/secret.html", SecureBaseURL))
	wds.verifySecretAuthorized(ctx, s.T())
}

// from network 192.168.240.202/32.
func (s *NetworkACLSuite) TestShouldAccessSecretUpon0FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wds, err := StartWebDriverWithProxy("http://proxy-client2.example.com:3128", GetWebDriverPort())
	s.Require().NoError(err)

	defer func() {
		err = wds.Stop()
		s.Require().NoError(err)
	}()

	wds.doVisit(s.T(), fmt.Sprintf("%s/secret.html", SecureBaseURL))
	wds.verifySecretAuthorized(ctx, s.T())
}

func TestNetworkACLSuite(t *testing.T) {
	suite.Run(t, NewNetworkACLSuite())
}
