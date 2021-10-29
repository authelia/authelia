package suites

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/poy/onpar"
)

func TestRunOIDCScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestOIDCScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldAuthorizeAccessToOIDCApp", func(t *testing.T, s RodSuite) {
			is := is.New(t)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doVisit(s.Context(ctx), OIDCBaseURL)
			s.verifyIsFirstFactorPage(t, s.Context(ctx))
			s.doFillLoginPageAndClick(t, s.Context(ctx), testUsername, testPassword, false)
			s.verifyIsSecondFactorPage(t, s.Context(ctx))
			s.doValidateTOTP(t, s.Context(ctx), secret)

			s.waitBodyContains(t, s.Context(ctx), "Not logged yet...")

			// Search for the 'login' link.
			err := s.Page.MustSearch("Log in").Click("left")
			is.NoErr(err)

			s.verifyIsConsentPage(t, s.Context(ctx))
			err = s.WaitElementLocatedByID(t, s.Context(ctx), "accept-button").Click("left")
			is.NoErr(err)

			// Verify that the app is showing the info related to the user stored in the JWT token.

			rUUID := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
			rInteger := regexp.MustCompile(`^\d+$`)
			rBoolean := regexp.MustCompile(`^(true|false)$`)
			rBase64 := regexp.MustCompile(`^[-_A-Za-z0-9+\\/]+([=]{0,3})$`)

			testCases := []struct {
				desc, elementID, elementText string
				pattern                      *regexp.Regexp
			}{
				{"welcome", "welcome", "Logged in as john!", nil},
				{"at_hash", "claim-at_hash", "", rBase64},
				{"jti", "claim-jti", "", rUUID},
				{"iat", "claim-iat", "", rInteger},
				{"nbf", "claim-nbf", "", rInteger},
				{"rat", "claim-rat", "", rInteger},
				{"expires", "claim-exp", "", rInteger},
				{"amr", "claim-amr", "pwd, otp, mfa", nil},
				{"acr", "claim-acr", "", nil},
				{"issuer", "claim-iss", "https://login.example.com:8080", nil},
				{"name", "claim-name", "John Doe", nil},
				{"preferred_username", "claim-preferred_username", "john", nil},
				{"groups", "claim-groups", "admins, dev", nil},
				{"email", "claim-email", "john.doe@authelia.com", nil},
				{"email_verified", "claim-email_verified", "", rBoolean},
			}

			var text string

			for _, tc := range testCases {
				t.Run(fmt.Sprintf("check_claims/%s", tc.desc), func(t *testing.T) {
					text, err = s.WaitElementLocatedByID(t, s.Context(ctx), tc.elementID).Text()
					is.NoErr(err)
					if tc.pattern == nil {
						is.Equal(tc.elementText, text)
					} else {
						is.True(tc.pattern.MatchString(text))
					}
				})
			}
		})

		o.Spec("TestShouldDenyConsent", func(t *testing.T, s RodSuite) {
			is := is.New(t)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doVisit(s.Context(ctx), OIDCBaseURL)
			s.verifyIsFirstFactorPage(t, s.Context(ctx))
			s.doFillLoginPageAndClick(t, s.Context(ctx), testUsername, testPassword, false)
			s.verifyIsSecondFactorPage(t, s.Context(ctx))
			s.doValidateTOTP(t, s.Context(ctx), secret)

			s.waitBodyContains(t, s.Context(ctx), "Not logged yet...")

			// Search for the 'login' link.
			err := s.Page.MustSearch("Log in").Click("left")
			is.NoErr(err)

			s.verifyIsConsentPage(t, s.Context(ctx))

			err = s.WaitElementLocatedByID(t, s.Context(ctx), "deny-button").Click("left")
			is.NoErr(err)

			s.verifyIsOIDC(t, s.Context(ctx), "access_denied", "https://oidc.example.com:8080/error?error=access_denied&error_description=The+resource+owner+or+authorization+server+denied+the+request.+Make+sure+that+the+request+you+are+making+is+valid.+Maybe+the+credential+or+request+parameters+you+are+using+are+limited+in+scope+or+otherwise+restricted.&state=random-string-here")

			errorDescription := "The resource owner or authorization server denied the request. Make sure that the request " +
				"you are making is valid. Maybe the credential or request parameters you are using are limited in scope or " +
				"otherwise restricted."

			s.verifyIsOIDCErrorPage(t, s.Context(ctx), "access_denied", errorDescription, "",
				"random-string-here")
		})
	})
}
