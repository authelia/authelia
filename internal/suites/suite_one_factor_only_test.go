package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/poy/onpar"
)

func TestOneFactorOnlySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
		s := setupTest(t, "", false)
		return t, s
	})

	o.AfterEach(func(t *testing.T, s RodSuite) {
		teardownTest(s)
	})

	// No target url is provided, then the user should be redirect to the default url.
	o.Spec("TestShouldRedirectUserToDefaultURL", func(t *testing.T, s RodSuite) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
		}()

		s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
		s.verifyIsHome(t, s.Context(ctx))
	})

	// Unsafe URL is provided, then the user should be redirect to the default url.
	o.Spec("TestShouldRedirectUserToDefaultURLWhenURLIsUnsafe", func(t *testing.T, s RodSuite) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
		}()

		s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "http://unsafe.local")
		s.verifyIsHome(t, s.Context(ctx))
	})

	// When use logged in and visit the portal again, she gets redirect to the authenticated view.
	o.Spec("TestShouldDisplayAuthenticatedView", func(t *testing.T, s RodSuite) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
		}()

		s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
		s.verifyIsHome(t, s.Context(ctx))
		s.doVisit(s.Context(ctx), GetLoginBaseURL())
		s.verifyIsAuthenticatedPage(t, s.Context(ctx))
	})

	o.Spec("TestShouldRedirectAlreadyAuthenticatedUser", func(t *testing.T, s RodSuite) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
		}()

		s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
		s.verifyIsHome(t, s.Context(ctx))

		s.doVisit(s.Context(ctx), fmt.Sprintf("%s?rd=https://singlefactor.example.com:8080/secret.html", GetLoginBaseURL()))
		s.verifySecretAuthorized(t, s.Context(ctx))
		s.verifyURLIs(t, s.Context(ctx), "https://singlefactor.example.com:8080/secret.html")
	})

	o.Spec("TestShouldNotRedirectAlreadyAuthenticatedUserToUnsafeURL", func(t *testing.T, s RodSuite) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer func() {
			cancel()
			s.collectScreenshot(ctx.Err(), s.Page)
		}()

		s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, "")
		s.verifyIsHome(t, s.Context(ctx))

		// Visit the login page and wait for redirection to 2FA page with success icon displayed.
		s.doVisit(s.Context(ctx), fmt.Sprintf("%s?rd=https://secure.example.local:8080", GetLoginBaseURL()))
		s.verifyNotificationDisplayed(t, s.Context(ctx), "Redirection was determined to be unsafe and aborted. Ensure the redirection URL is correct.")
	})
}
