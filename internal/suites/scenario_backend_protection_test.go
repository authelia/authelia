package suites

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/matryer/is"
	"github.com/poy/onpar"
)

func TestRunBackendProtectionScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	// WARNING: This scenario should be run with TLS enabled in the authelia backend.
	o.Group("TestBackendProtectionScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestProtectionOfBackendEndpoints", func(t *testing.T, s RodSuite) {
			AssertRequestStatusCode(t, "POST", fmt.Sprintf("%s/api/secondfactor/totp", AutheliaBaseURL), 403)
			AssertRequestStatusCode(t, "POST", fmt.Sprintf("%s/api/secondfactor/webauthn/assertion", AutheliaBaseURL), 403)
			AssertRequestStatusCode(t, "POST", fmt.Sprintf("%s/api/secondfactor/webauthn/attestation", AutheliaBaseURL), 403)
			AssertRequestStatusCode(t, "POST", fmt.Sprintf("%s/api/user/info/2fa_method", AutheliaBaseURL), 403)

			AssertRequestStatusCode(t, "GET", fmt.Sprintf("%s/api/user/info", AutheliaBaseURL), 403)
			AssertRequestStatusCode(t, "GET", fmt.Sprintf("%s/api/configuration", AutheliaBaseURL), 403)

			AssertRequestStatusCode(t, "POST", fmt.Sprintf("%s/api/secondfactor/totp/identity/start", AutheliaBaseURL), 403)
			AssertRequestStatusCode(t, "POST", fmt.Sprintf("%s/api/secondfactor/totp/identity/finish", AutheliaBaseURL), 403)
			AssertRequestStatusCode(t, "POST", fmt.Sprintf("%s/api/secondfactor/webauthn/identity/start", AutheliaBaseURL), 403)
			AssertRequestStatusCode(t, "POST", fmt.Sprintf("%s/api/secondfactor/webauthn/identity/finish", AutheliaBaseURL), 403)
		})

		o.Spec("TestInvalidEndpointsReturn404", func(t *testing.T, s RodSuite) {
			AssertRequestStatusCode(t, "GET", fmt.Sprintf("%s/api/not_existing", AutheliaBaseURL), 404)
			AssertRequestStatusCode(t, "HEAD", fmt.Sprintf("%s/api/not_existing", AutheliaBaseURL), 404)
			AssertRequestStatusCode(t, "POST", fmt.Sprintf("%s/api/not_existing", AutheliaBaseURL), 404)

			AssertRequestStatusCode(t, "GET", fmt.Sprintf("%s/api/not_existing/second", AutheliaBaseURL), 404)
			AssertRequestStatusCode(t, "HEAD", fmt.Sprintf("%s/api/not_existing/second", AutheliaBaseURL), 404)
			AssertRequestStatusCode(t, "POST", fmt.Sprintf("%s/api/not_existing/second", AutheliaBaseURL), 404)
		})

		o.Spec("TestInvalidEndpointsReturn405", func(t *testing.T, s RodSuite) {
			AssertRequestStatusCode(t, "PUT", fmt.Sprintf("%s/api/configuration", AutheliaBaseURL), 405)
		})
	})
}

func AssertRequestStatusCode(t *testing.T, method, url string, expectedStatusCode int) {
	t.Run(url, func(t *testing.T) {
		is := is.New(t)
		req, err := http.NewRequest(method, url, nil)
		is.NoErr(err)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Needs to be enabled in suites. Not used in production.
		}
		client := &http.Client{
			Transport: tr,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		res, err := client.Do(req)
		is.NoErr(err)
		is.Equal(expectedStatusCode, res.StatusCode)
	})
}
