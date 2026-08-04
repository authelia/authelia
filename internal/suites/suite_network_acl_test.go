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

const newIPEmailSubject = "[Authelia] Login From New IP"

// These known-IP / new-IP-notification tests only exercise the API layer (login, logout, notification emails), so
// they talk to Authelia directly through the suite's squid proxies instead of driving a full browser - this avoids
// the overhead/flakiness of spinning up a headless Chrome session and loading the frontend just to submit a form.

// TestShouldSendEmailOnLoginFromNewIP logs in from 192.168.240.201/32 (client-1) as a user who has never
// authenticated in this suite run, and expects a "Login From New IP" notification referencing that IP.
func (s *NetworkACLSuite) TestShouldSendEmailOnLoginFromNewIP() {
	client := newAPIClientWithProxy(s.T(), "http://proxy-client1.example.com:3128")

	apiFirstFactorLogin(s.T(), client, "harry", "password")

	msg := doGetLastEmailMessageWithSubjectForRecipient(s.T(), newIPEmailSubject, "harry.potter@authelia.com")

	content, err := msg.GetContent()
	s.Require().NoError(err)
	s.Assert().Contains(string(content), "192.168.240.201")
}

// TestShouldNotSendEmailOnLoginFromKnownIP logs in twice from the same IP (192.168.240.201/32) and expects only
// the first login to trigger a notification.
func (s *NetworkACLSuite) TestShouldNotSendEmailOnLoginFromKnownIP() {
	client := newAPIClientWithProxy(s.T(), "http://proxy-client1.example.com:3128")

	apiFirstFactorLogin(s.T(), client, "bob", "password")

	// Consume (and thereby mark read) the notification triggered by this first login, so the check below can't
	// pass by accidentally matching it.
	msg := doGetLastEmailMessageWithSubjectForRecipient(s.T(), newIPEmailSubject, "bob.dylan@authelia.com")
	_, err := msg.GetContent()
	s.Require().NoError(err)

	apiLogout(s.T(), client)

	apiFirstFactorLogin(s.T(), client, "bob", "password")

	s.Assert().False(doHasUnreadEmailMessageWithSubjectForRecipient(s.T(), newIPEmailSubject, "bob.dylan@authelia.com"),
		"a second login from the same known IP should not send another notification")
}

// TestShouldSendSeparateEmailForDifferentIP logs in as the same user first from 192.168.240.201/32 (client-1) then
// from 192.168.240.202/32 (client-2), and expects a distinct notification for each IP.
func (s *NetworkACLSuite) TestShouldSendSeparateEmailForDifferentIP() {
	client1 := newAPIClientWithProxy(s.T(), "http://proxy-client1.example.com:3128")

	apiFirstFactorLogin(s.T(), client1, "james", "password")

	msgA := doGetLastEmailMessageWithSubjectForRecipient(s.T(), newIPEmailSubject, "james.dean@authelia.com")
	contentA, err := msgA.GetContent()
	s.Require().NoError(err)
	s.Assert().Contains(string(contentA), "192.168.240.201")

	client2 := newAPIClientWithProxy(s.T(), "http://proxy-client2.example.com:3128")

	apiFirstFactorLogin(s.T(), client2, "james", "password")

	msgB := doGetLastEmailMessageWithSubjectForRecipient(s.T(), newIPEmailSubject, "james.dean@authelia.com")
	contentB, err := msgB.GetContent()
	s.Require().NoError(err)
	s.Assert().Contains(string(contentB), "192.168.240.202")
}

func TestNetworkACLSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewNetworkACLSuite())
}
