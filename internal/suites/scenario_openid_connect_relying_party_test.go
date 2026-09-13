// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	openIDConnectRelyingPartyProviderID    = "upstream"
	openIDConnectRelyingPartyProviderName  = "Upstream"
	openIDConnectRelyingPartyCallbackPath  = "/api/identity/upstream/callback"
	openIDConnectRelyingPartySignInButton  = "external-identity-sign-in-button-upstream"
	openIDConnectRelyingPartyLinkButton    = "external-identity-link-start-upstream"
	openIDConnectRelyingPartyPendingPanel  = "external-identity-link-pending-panel"
	openIDConnectRelyingPartyLinksPanel    = "external-identity-links-panel"
	openIDConnectRelyingPartyLinkNotice    = "external-identity-link-notice"
	openIDConnectRelyingPartyLinkSelector  = `[id^='external-identity-link-'][id$='-description']`
	openIDConnectRelyingPartyDeleteButton  = `[id^='external-identity-link-'][id$='-delete']`
	openIDConnectRelyingPartyResumeQueryFm = "%s/external-identity/link?link_provider=%s"
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

	s.doStartExternalIdentityLogin(s.T(), page, "")
	s.doAuthenticateAtProvider(s.T(), page, "john")

	s.verifyURLIs(s.T(), page, fmt.Sprintf(openIDConnectRelyingPartyResumeQueryFm, LoginBaseURL, openIDConnectRelyingPartyProviderID))
	s.verifyIsFirstFactorPage(s.T(), page)
	require.True(s.T(), s.CheckElementExistsLocatedByID(s.T(), page, openIDConnectRelyingPartyLinkNotice))
	require.False(s.T(), s.CheckElementExistsLocatedByID(s.T(), page, openIDConnectRelyingPartySignInButton))

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

	s.doStartExternalIdentityLogin(s.T(), page, target)
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

	s.ClickElementLocatedByID(s.T(), page, "external-identity-link-decline")

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

	s.doStartExternalIdentityLogin(s.T(), page, "")

	callback := s.doCaptureCallbackURL(ctx, page, func() {
		s.doAuthenticateAtProvider(s.T(), page, "mary")
	})

	require.Contains(s.T(), callback, openIDConnectRelyingPartyCallbackPath)
	s.verifyIsAuthenticatedPage(s.T(), page)

	s.doLogout(s.T(), page)

	s.doVisit(s.T(), page, callback)

	s.verifyIsFirstFactorPage(s.T(), page)
	s.verifyURLIs(s.T(), page, fmt.Sprintf("%s/?external_identity_error=true", LoginBaseURL))
	s.verifyNotificationDisplayed(s.T(), page, "There was an issue signing in with the external provider")
}

func (s *OpenIDConnectRelyingPartyScenario) doStartExternalIdentityLogin(t *testing.T, page *rod.Page, target string) {
	s.doVisitLoginPage(t, page, BaseDomain, target)
	s.ClickElementLocatedByID(t, page, openIDConnectRelyingPartySignInButton)
}

func (s *OpenIDConnectRelyingPartyScenario) doStartLinkFromSettings(t *testing.T, page *rod.Page) {
	s.doVisit(t, page, LinkedAccountsURL)
	s.WaitElementLocatedByID(t, page, openIDConnectRelyingPartyLinksPanel)
	s.ClickElementLocatedByID(t, page, openIDConnectRelyingPartyLinkButton)
}

func (s *OpenIDConnectRelyingPartyScenario) doAuthenticateAtProvider(t *testing.T, page *rod.Page, username string) {
	s.waitURLHasPrefix(t, page, UpstreamBaseURL)
	s.verifyIsFirstFactorPage(t, page)
	s.doFillLoginPageAndClick(t, page, username, testPassword, false)

	// Where the browser lands depends on the callback: the portal for a link that does not exist, the target for one
	// that does. All that is known here is that it leaves the provider.
	s.waitURLLacksPrefix(t, page, UpstreamBaseURL)
}

func (s *OpenIDConnectRelyingPartyScenario) doLogoutProvider(t *testing.T, page *rod.Page) {
	s.doVisit(t, page, fmt.Sprintf("%s/logout", UpstreamBaseURL))
	s.verifyIsFirstFactorPage(t, page)
}

func (s *OpenIDConnectRelyingPartyScenario) doAcceptPendingLink(t *testing.T, page *rod.Page) {
	s.WaitElementLocatedByID(t, page, openIDConnectRelyingPartyPendingPanel)

	s.ClickElementLocatedByID(t, page, "external-identity-link-accept")

	s.WaitElementLocatedByID(t, page, "dialog-verify-one-time-code")
	s.doMustVerifyIdentity(t, page)

	s.verifyNotificationDisplayed(t, page, fmt.Sprintf("Successfully linked the %s account", openIDConnectRelyingPartyProviderName))
}

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

func (s *OpenIDConnectRelyingPartyScenario) doCaptureCallbackURL(ctx context.Context, page *rod.Page, action func()) string {
	callback, _, _, err := s.doAwaitExternalIdentityCallback(ctx, page, openIDConnectRelyingPartyProviderID, elementLocateTimeout, func() error {
		action()

		return nil
	})

	require.NoError(s.T(), err)

	return callback
}

func TestRunOpenIDConnectRelyingPartyScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOpenIDConnectRelyingPartyScenario())
}
