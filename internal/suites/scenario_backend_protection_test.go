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
		s.Assert().Equal(res.StatusCode, expectedStatusCode)
	})
}

func (s *BackendProtectionScenario) TestProtectionOfBackendEndpoints() {
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/totp", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/u2f/sign", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/u2f/register", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/u2f/sign_request", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/user/info/2fa_method", AutheliaBaseURL), 403)

	s.AssertRequestStatusCode("GET", fmt.Sprintf("%s/api/user/info", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("GET", fmt.Sprintf("%s/api/configuration", AutheliaBaseURL), 403)

	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/u2f/identity/start", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/u2f/identity/finish", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/totp/identity/start", AutheliaBaseURL), 403)
	s.AssertRequestStatusCode("POST", fmt.Sprintf("%s/api/secondfactor/totp/identity/finish", AutheliaBaseURL), 403)
}

func TestRunBackendProtection(t *testing.T) {
	suite.Run(t, NewBackendProtectionScenario())
}
