package suites

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

// WARNING: This scenario is intended to be used with TLS enabled in the authelia backend.

type BackendProtectionScenario struct {
	suite.Suite
}

func NewBackendProtectionScenario() *BackendProtectionScenario {
	return &BackendProtectionScenario{}
}

func (s *BackendProtectionScenario) AssertRequestStatusCode(method, url string, expectedStatusCode int) {
	s.Run(url, func() {
		req, err := http.NewRequest(method, url, nil)
		s.Assert().NoError(err)

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
		s.Assert().NoError(err)
		s.Assert().Equal(expectedStatusCode, res.StatusCode)
	})
}

func (s *BackendProtectionScenario) TestProtectionOfBackendEndpoints() {
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/totp", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("GET", fmt.Sprintf("%s/api/secondfactor/webauthn/credentials", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/webauthn", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("PUT", fmt.Sprintf("%s/api/secondfactor/webauthn/credential/register", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/webauthn/credential/register", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("DELETE", fmt.Sprintf("%s/api/secondfactor/webauthn/credential/1", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("PUT", fmt.Sprintf("%s/api/secondfactor/webauthn/credential/1", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/user/info/2fa_method", AutheliaBaseURL), 403)

	s.AssertRequestStatusCode("GET", fmt.Sprintf("%s/api/user/info", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("GET", fmt.Sprintf("%s/api/configuration", AutheliaBaseURL), 403)

	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/totp/identity/start", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/totp/identity/finish", AutheliaBaseURL), 403)
}

func (s *BackendProtectionScenario) TestInvalidEndpointsReturn404() {
	s.AssertRequestStatusCode("GET", fmt.Sprintf("%s/api/not_existing", AutheliaBaseURL), 404)
	s.AssertRequestStatusCode("HEAD", fmt.Sprintf("%s/api/not_existing", AutheliaBaseURL), 404)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/not_existing", AutheliaBaseURL), 404)

	s.AssertRequestStatusCode("GET", fmt.Sprintf("%s/api/not_existing/second", AutheliaBaseURL), 404)
	s.AssertRequestStatusCode("HEAD", fmt.Sprintf("%s/api/not_existing/second", AutheliaBaseURL), 404)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/not_existing/second", AutheliaBaseURL), 404)
}

func (s *BackendProtectionScenario) TestInvalidEndpointsReturn405() {
	s.AssertRequestStatusCode("PUT", fmt.Sprintf("%s/api/configuration", AutheliaBaseURL), 405)
}

func TestRunBackendProtection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewBackendProtectionScenario())
}
