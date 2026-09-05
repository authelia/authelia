package suites

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	openIDConnectRelyingPartyProviderID    = "upstream"
	openIDConnectRelyingPartyProviderName  = "Upstream"
	openIDConnectRelyingPartyCallbackPath  = "/api/firstfactor/openid-connect/upstream/callback"
	openIDConnectRelyingPartySignInButton  = "openid-connect-sign-in-button-upstream"
	openIDConnectRelyingPartyLinkButton    = "openid-connect-link-start-upstream"
	openIDConnectRelyingPartyPendingPanel  = "openid-connect-link-pending-panel"
	openIDConnectRelyingPartyLinksPanel    = "openid-connect-links-panel"
	openIDConnectRelyingPartyLinkSelector  = `[id^='openid-connect-link-'][id$='-description']`
	openIDConnectRelyingPartyDeleteButton  = `[id^='openid-connect-link-'][id$='-delete']`
	openIDConnectRelyingPartyResumeQueryFm = "%s/?link_provider=%s"
)

type OpenIDConnectRelyingPartyScenario struct {
	*RodSuite
}

func NewOpenIDConnectRelyingPartyScenario() *OpenIDConnectRelyingPartyScenario {
	return &OpenIDConnectRelyingPartyScenario{
		RodSuite: NewRodSuite(""),
	}
}

func (s *OpenIDConnectRelyingPartyScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *OpenIDConnectRelyingPartyScenario) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *OpenIDConnectRelyingPartyScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
}

func (s *OpenIDConnectRelyingPartyScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *OpenIDConnectRelyingPartyScenario) TestShouldLinkFromUnauthenticatedStart() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	page := s.Context(ctx)

	s.doStartOpenIDConnectLogin(s.T(), page, "")
	s.doAuthenticateAtProvider(s.T(), page, "john")

	s.verifyURLIs(s.T(), page, fmt.Sprintf(openIDConnectRelyingPartyResumeQueryFm, LoginBaseURL, openIDConnectRelyingPartyProviderID))
	s.verifyIsFirstFactorPage(s.T(), page)

	s.doFillLoginPageAndClick(s.T(), page, "john", testPassword, false)

	s.waitURLHasPrefix(s.T(), page, LinkedAccountsURL)

	s.doAcceptPendingLink(s.T(), page)
	s.verifyLinkIsListed(s.T(), page)
}

func (s *OpenIDConnectRelyingPartyScenario) TestShouldLinkFromAuthenticatedSession() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	page := s.Context(ctx)

	s.doLoginOneFactor(s.T(), page, "harry", testPassword, false, BaseDomain, "")
	s.verifyIsAuthenticatedPage(s.T(), page)

	s.doStartLinkFromSettings(s.T(), page)
	s.doAuthenticateAtProvider(s.T(), page, "harry")

	s.verifyURLIs(s.T(), page, LinkedAccountsURL)

	s.doAcceptPendingLink(s.T(), page)
	s.verifyLinkIsListed(s.T(), page)
}

func (s *OpenIDConnectRelyingPartyScenario) TestShouldLoginWithExistingLink() {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	page := s.Context(ctx)

	s.doLinkAccount(s.T(), page, "bob")

	s.doLogoutProvider(s.T(), page)
	s.doLogout(s.T(), page)

	target := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)

	s.doStartOpenIDConnectLogin(s.T(), page, target)
	s.doAuthenticateAtProvider(s.T(), page, "bob")

	s.verifyURLIs(s.T(), page, target)
	s.verifySecretAuthorized(s.T(), page)
}

func (s *OpenIDConnectRelyingPartyScenario) TestShouldDeclinePendingLink() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	page := s.Context(ctx)

	s.doLoginOneFactor(s.T(), page, "james", testPassword, false, BaseDomain, "")
	s.verifyIsAuthenticatedPage(s.T(), page)

	s.doStartLinkFromSettings(s.T(), page)
	s.doAuthenticateAtProvider(s.T(), page, "james")

	s.WaitElementLocatedByID(s.T(), page, openIDConnectRelyingPartyPendingPanel)

	s.ClickElementLocatedByID(s.T(), page, "openid-connect-link-decline")

	s.verifyNotificationDisplayed(s.T(), page, fmt.Sprintf("Successfully declined the %s account", openIDConnectRelyingPartyProviderName))

	s.doVisit(s.T(), page, LinkedAccountsURL)
	s.WaitElementLocatedByID(s.T(), page, openIDConnectRelyingPartyLinksPanel)

	require.False(s.T(), s.CheckElementExistsLocatedByID(s.T(), page, openIDConnectRelyingPartyPendingPanel))
	require.False(s.T(), s.CheckElementExistsLocatedBySelector(s.T(), page, openIDConnectRelyingPartyLinkSelector))

	s.WaitElementLocatedByID(s.T(), page, openIDConnectRelyingPartyLinkButton)
}

func (s *OpenIDConnectRelyingPartyScenario) TestShouldDeleteExistingLink() {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	page := s.Context(ctx)

	s.doLinkAccount(s.T(), page, "alice")

	s.ClickElementLocatedBySelector(s.T(), page, openIDConnectRelyingPartyDeleteButton)

	s.doMaybeVerifyIdentity(s.T(), page, "#dialog-delete")

	s.ClickElementLocatedByID(s.T(), page, "dialog-delete")

	s.verifyNotificationDisplayed(s.T(), page, fmt.Sprintf("Successfully deleted the %s account", openIDConnectRelyingPartyProviderName))

	s.doVisit(s.T(), page, LinkedAccountsURL)
	s.WaitElementLocatedByID(s.T(), page, openIDConnectRelyingPartyLinksPanel)

	require.False(s.T(), s.CheckElementExistsLocatedBySelector(s.T(), page, openIDConnectRelyingPartyLinkSelector))

	s.WaitElementLocatedByID(s.T(), page, openIDConnectRelyingPartyLinkButton)
}

func (s *OpenIDConnectRelyingPartyScenario) TestShouldRejectReplayedState() {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	page := s.Context(ctx)

	s.doLinkAccount(s.T(), page, "mary")

	s.doLogoutProvider(s.T(), page)
	s.doLogout(s.T(), page)

	s.doStartOpenIDConnectLogin(s.T(), page, "")

	callback := s.doCaptureCallbackURL(ctx, page, func() {
		s.doAuthenticateAtProvider(s.T(), page, "mary")
	})

	require.Contains(s.T(), callback, openIDConnectRelyingPartyCallbackPath)
	s.verifyIsAuthenticatedPage(s.T(), page)

	s.doLogout(s.T(), page)

	s.doVisit(s.T(), page, callback)

	s.verifyIsFirstFactorPage(s.T(), page)
	s.verifyURLIs(s.T(), page, fmt.Sprintf("%s/", LoginBaseURL))
}

// doStartOpenIDConnectLogin visits the login portal and clicks the button for the external provider. The target is the
// resource the user is trying to reach, and is empty when they visit the portal directly.
func (s *OpenIDConnectRelyingPartyScenario) doStartOpenIDConnectLogin(t *testing.T, page *rod.Page, target string) {
	s.doVisitLoginPage(t, page, BaseDomain, target)
	s.ClickElementLocatedByID(t, page, openIDConnectRelyingPartySignInButton)
}

// doStartLinkFromSettings starts a linking flow from the Linked Accounts settings page, which is the entry point an
// already authenticated user has.
func (s *OpenIDConnectRelyingPartyScenario) doStartLinkFromSettings(t *testing.T, page *rod.Page) {
	s.doVisit(t, page, LinkedAccountsURL)
	s.WaitElementLocatedByID(t, page, openIDConnectRelyingPartyLinksPanel)
	s.ClickElementLocatedByID(t, page, openIDConnectRelyingPartyLinkButton)
}

// doAuthenticateAtProvider completes the first factor at the external provider and waits for the browser to be
// returned to the instance under test.
func (s *OpenIDConnectRelyingPartyScenario) doAuthenticateAtProvider(t *testing.T, page *rod.Page, username string) {
	s.waitURLHasPrefix(t, page, UpstreamBaseURL)
	s.verifyIsFirstFactorPage(t, page)
	s.doFillLoginPageAndClick(t, page, username, testPassword, false)

	// Where the browser lands depends on the callback: the portal for a link that does not exist, the target for one
	// that does. All that is known here is that it leaves the provider.
	s.waitURLLacksPrefix(t, page, UpstreamBaseURL)
}

// doLogoutProvider ends the session at the external provider. The provider does not prompt a user it still has a
// session for, so a scenario which runs a second flow in the same browser clears it first to keep the prompt of the
// second flow the same as the prompt of the first.
func (s *OpenIDConnectRelyingPartyScenario) doLogoutProvider(t *testing.T, page *rod.Page) {
	s.doVisit(t, page, fmt.Sprintf("%s/logout", UpstreamBaseURL))
	s.verifyIsFirstFactorPage(t, page)
}

// doAcceptPendingLink accepts the proposal shown on the Linked Accounts settings page. The accept is behind an
// elevated session, so the identity verification the elevation requires is completed here rather than routed around.
func (s *OpenIDConnectRelyingPartyScenario) doAcceptPendingLink(t *testing.T, page *rod.Page) {
	s.WaitElementLocatedByID(t, page, openIDConnectRelyingPartyPendingPanel)

	s.ClickElementLocatedByID(t, page, "openid-connect-link-accept")

	s.WaitElementLocatedByID(t, page, "dialog-verify-one-time-code")
	s.doMustVerifyIdentity(t, page)

	s.verifyNotificationDisplayed(t, page, fmt.Sprintf("Successfully linked the %s account", openIDConnectRelyingPartyProviderName))
}

// doLinkAccount is the fixture for a scenario which needs an existing link. It signs the user in and links through the
// settings entry point, which prompts at the provider every time and so does not depend on a session there.
func (s *OpenIDConnectRelyingPartyScenario) doLinkAccount(t *testing.T, page *rod.Page, username string) {
	s.doLoginOneFactor(t, page, username, testPassword, false, BaseDomain, "")
	s.verifyIsAuthenticatedPage(t, page)

	s.doStartLinkFromSettings(t, page)
	s.doAuthenticateAtProvider(t, page, username)

	s.doAcceptPendingLink(t, page)
	s.verifyLinkIsListed(t, page)
}

func (s *OpenIDConnectRelyingPartyScenario) verifyLinkIsListed(t *testing.T, page *rod.Page) {
	s.WaitElementLocatedByID(t, page, openIDConnectRelyingPartyLinksPanel)

	element := s.WaitElementLocatedBySelector(t, page, openIDConnectRelyingPartyLinkSelector)

	text, err := element.Text()
	require.NoError(t, err)
	require.Equal(t, openIDConnectRelyingPartyProviderName, text)

	require.False(t, s.CheckElementExistsLocatedByID(t, page, openIDConnectRelyingPartyLinkButton))
}

// waitURLHasPrefix waits until the page is at a URL beginning with the prefix. A redirect chain does not commit a
// navigation for each hop, so waiting on an element alone cannot tell which of the two portals is being looked at.
func (s *OpenIDConnectRelyingPartyScenario) waitURLHasPrefix(t *testing.T, page *rod.Page, prefix string) {
	s.waitURL(t, page, prefix, true)
}

// waitURLLacksPrefix waits until the page is at a URL which does not begin with the prefix.
func (s *OpenIDConnectRelyingPartyScenario) waitURLLacksPrefix(t *testing.T, page *rod.Page, prefix string) {
	s.waitURL(t, page, prefix, false)
}

func (s *OpenIDConnectRelyingPartyScenario) waitURL(t *testing.T, page *rod.Page, prefix string, has bool) {
	var current string

	for {
		if err := page.GetContext().Err(); err != nil {
			require.FailNowf(t, "timeout waiting for the page to navigate",
				"the page did not reach a url which %s the prefix '%s', it is at '%s'",
				map[bool]string{true: "has", false: "lacks"}[has], prefix, current)

			return
		}

		if info, err := page.Info(); err == nil {
			current = info.URL

			if strings.HasPrefix(current, prefix) == has {
				return
			}
		}

		time.Sleep(elementRetryInterval)
	}
}

// doCaptureCallbackURL records the redirect URI request the external provider sends the browser to. The hops of a
// redirect chain are requests rather than committed navigations, so the URL is taken from the network event.
func (s *OpenIDConnectRelyingPartyScenario) doCaptureCallbackURL(ctx context.Context, page *rod.Page, action func()) string {
	var captured string

	wait := page.Context(ctx).EachEvent(func(e *proto.NetworkRequestWillBeSent) bool {
		if !strings.Contains(e.Request.URL, openIDConnectRelyingPartyCallbackPath) {
			return false
		}

		captured = e.Request.URL

		return true
	})

	action()

	wait()

	return captured
}

func TestRunOpenIDConnectRelyingPartyScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOpenIDConnectRelyingPartyScenario())
}
