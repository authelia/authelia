package suites

import (
	"context"
	"testing"
	"time"

	"github.com/poy/onpar"
)

var redirectionAuthorizations = map[string]bool{
	// external website.
	"https://www.google.fr": false,
	// Not the right domain.
	"https://public.example.com.a:8080/secret.html": false,
	// Not https.
	"http://secure.example.com:8080/secret.html": false,
	// Domain handled by Authelia.
	"https://secure.example.com:8080/secret.html": true,
}

var logoutRedirectionURLs = map[string]bool{
	// external website.
	"https://www.google.fr": false,
	// Not the right domain.
	"https://public.example-not-right.com:8080/index.html": false,
	// Not https.
	"http://public.example.com:8080/index.html": false,
	// Domain handled by Authelia.
	"https://public.example.com:8080/index.html": true,
}

func TestRunRedirectionCheckScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestRedirectionCheckScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldRedirectOnLoginOnlyWhenDomainIsSafe", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			for url, redirected := range redirectionAuthorizations {
				t.Run(url, func(t *testing.T) {
					s.doLoginTwoFactor(t, s.Context(ctx), testUsername, testPassword, false, secret, url)

					if redirected {
						s.verifySecretAuthorized(t, s.Context(ctx))
					} else {
						s.verifyIsAuthenticatedPage(t, s.Context(ctx))
					}

					s.doLogout(t, s.Context(ctx))
				})
			}
		})

		o.Spec("TestShouldRedirectOnLogoutOnlyWhenDomainIsSafe", func(t *testing.T, s RodSuite) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			for url, success := range logoutRedirectionURLs {
				t.Run(url, func(t *testing.T) {
					s.doLogoutWithRedirect(t, s.Context(ctx), url, !success)
				})
			}
		})
	})
}
