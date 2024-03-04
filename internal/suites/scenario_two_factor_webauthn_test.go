package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TwoFactorWebAuthnSuite struct {
	*RodSuite
}

func NewTwoFactorWebAuthnScenario() *TwoFactorWebAuthnSuite {
	return &TwoFactorWebAuthnSuite{
		RodSuite: NewRodSuite(""),
	}
}

func (s *TwoFactorWebAuthnSuite) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)

		s.collectCoverage(s.Page)
		s.MustClose()
	}()

	s.Page = s.doCreateTab(s.T(), HomeBaseURL)

	s.doWebAuthnInitialize(s.T(), s.Page, false)

	s.doLoginAndRegisterWebAuthn(s.T(), s.Context(ctx), "john", "password", false)
	s.doLogout(s.T(), s.Page)
}

func (s *TwoFactorWebAuthnSuite) TearDownSuite() {
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *TwoFactorWebAuthnSuite) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)

	s.doWebAuthnInitialize(s.T(), s.Page, false)
	s.doWebAuthnRestoreCredentials(s.T(), s.Page)
}

func (s *TwoFactorWebAuthnSuite) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *TwoFactorWebAuthnSuite) TestShouldAuthorizeSecretAfterTwoFactor() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	username := testUsername
	password := testPassword

	targetURL := fmt.Sprintf("%s/secret.html", AdminBaseURL)
	s.doLoginOneFactor(s.T(), s.Context(ctx), username, password, false, BaseDomain, targetURL)

	s.doWebAuthnMethodMaybeSelect(s.T(), s.Context(ctx))

	// And check if the user is redirected to the secret.
	s.verifySecretAuthorized(s.T(), s.Context(ctx))

	// Leave the secret.
	s.doVisit(s.T(), s.Context(ctx), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Context(ctx))

	// And try to reload it again to check the session is kept.
	s.doVisit(s.T(), s.Context(ctx), targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}

func (s *TwoFactorWebAuthnSuite) TestShouldRenameCredential() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	username := testUsername
	password := testPassword

	s.doLoginOneFactor(s.T(), s.Context(ctx), username, password, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.doOpenSettingsMenuClickTwoFactor(s.T(), s.Context(ctx))

	s.Assert().Equal("testing", s.WaitElementLocatedByID(s.T(), s.Context(ctx), "webauthn-credential-0-description").MustText())

	s.doWebAuthnCredentialRename(s.T(), s.Context(ctx), "testing2")

	s.Assert().Equal("testing2", s.WaitElementLocatedByID(s.T(), s.Context(ctx), "webauthn-credential-0-description").MustText())

	s.doWebAuthnCredentialRename(s.T(), s.Context(ctx), "testing")

	s.Assert().Equal("testing", s.WaitElementLocatedByID(s.T(), s.Context(ctx), "webauthn-credential-0-description").MustText())
}

func (s *TwoFactorWebAuthnSuite) TestShouldShowCredentialInformation() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	username := testUsername
	password := testPassword

	s.doLoginOneFactor(s.T(), s.Context(ctx), username, password, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.doOpenSettingsMenuClickTwoFactor(s.T(), s.Context(ctx))

	s.Require().NoError(s.WaitElementLocatedByID(s.T(), s.Context(ctx), "webauthn-credential-0-information").Click("left", 1))
	s.Require().NoError(s.WaitElementLocatedByID(s.T(), s.Context(ctx), "dialog-close").Click("left", 1))
}

func (s *TwoFactorWebAuthnSuite) TestShouldDeleteAndRegisterCredential() {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	username := testUsername
	password := testPassword

	s.doLoginOneFactor(s.T(), s.Context(ctx), username, password, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.doOpenSettingsMenuClickTwoFactor(s.T(), s.Context(ctx))

	s.doWebAuthnCredentialMustDelete(s.T(), s.Context(ctx))
	s.doWebAuthnCredentialRegister(s.T(), s.Context(ctx), "testing")
}

func TestRunTwoFactorWebAuthn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewTwoFactorWebAuthnScenario())
}
