package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type NetworkACLSuite struct {
	*BaseSuite
}

func NewNetworkACLSuite() *NetworkACLSuite {
	return &NetworkACLSuite{
		BaseSuite: &BaseSuite{
			Name: networkACLSuiteName,
		},
	}
}

func (s *NetworkACLSuite) TestShouldAccessSecretUpon2FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	browser, err := NewRodSession()
	s.Require().NoError(err)

	defer func() {
		err = browser.WebDriver.Close()
		s.Require().NoError(err)
		browser.Launcher.Cleanup()
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SecureBaseURL)
	page := browser.doCreateTab(s.T(), targetURL).Context(ctx)

	browser.verifyIsFirstFactorPage(s.T(), page)
	browser.doRegisterTOTPAndLogin2FA(s.T(), page, "john", "password", false, targetURL)
	browser.verifySecretAuthorized(s.T(), page)
}

// from network 192.168.240.201/32.
func (s *NetworkACLSuite) TestShouldAccessSecretUpon1FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	browser, err := NewRodSession(RodSessionWithProxy("http://proxy-client1.example.com:3128"))
	s.Require().NoError(err)

	defer func() {
		err = browser.WebDriver.Close()
		s.Require().NoError(err)
		browser.Launcher.Cleanup()
	}()

	targetURL := fmt.Sprintf("%s/secret.html", SecureBaseURL)
	page := browser.doCreateTab(s.T(), targetURL).Context(ctx)

	browser.verifyIsFirstFactorPage(s.T(), page)
	browser.doLoginOneFactor(s.T(), page, "john", "password",
		false, BaseDomain, fmt.Sprintf("%s/secret.html", SecureBaseURL))
	browser.verifySecretAuthorized(s.T(), page)
}

// from network 192.168.240.202/32.
func (s *NetworkACLSuite) TestShouldAccessSecretUpon0FA() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	browser, err := NewRodSession(RodSessionWithProxy("http://proxy-client2.example.com:3128"))
	s.Require().NoError(err)

	defer func() {
		err = browser.WebDriver.Close()
		s.Require().NoError(err)
		browser.Launcher.Cleanup()
	}()

	page := browser.doCreateTab(s.T(), fmt.Sprintf("%s/secret.html", SecureBaseURL)).Context(ctx)

	browser.verifySecretAuthorized(s.T(), page)
}

func TestNetworkACLSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewNetworkACLSuite())
}
