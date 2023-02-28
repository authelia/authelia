package suites

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

func NewRequestMethodScenario() *RequestMethodScenario {
	return &RequestMethodScenario{}
}

type RequestMethodScenario struct {
	suite.Suite

	client *http.Client
}

func (s *RequestMethodScenario) SetupSuite() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Needs to be enabled in suites. Not used in production.
	}

	s.client = &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (s *RequestMethodScenario) TestShouldRespondWithAppropriateMethodNotAllowedHeaders() {
	testCases := []struct {
		name     string
		method   string
		uri      string
		expected []string
	}{
		{"RootPathShouldShowAllowedMethodsOnInvalidRequest", fasthttp.MethodPost, AutheliaBaseURL, []string{fasthttp.MethodGet, fasthttp.MethodHead, fasthttp.MethodOptions}},
		{"OpenAPISpecificationShouldShowAllowedMethodsOnInvalidRequest", fasthttp.MethodPost, fmt.Sprintf("%s/api/openapi.yml", AutheliaBaseURL), []string{fasthttp.MethodGet, fasthttp.MethodHead, fasthttp.MethodOptions}},
		{"LocalesShouldShowAllowedMethodsOnInvalidRequest", fasthttp.MethodPost, fmt.Sprintf("%s/locales/en/portal.json", AutheliaBaseURL), []string{fasthttp.MethodGet, fasthttp.MethodHead, fasthttp.MethodOptions}},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := http.NewRequest(tc.method, tc.uri, nil)
			s.Assert().NoError(err)

			res, err := s.client.Do(req)

			s.Assert().NoError(err)
			s.Assert().Equal(fasthttp.StatusMethodNotAllowed, res.StatusCode)
			s.Assert().Equal(strings.Join(tc.expected, ", "), res.Header.Get(fasthttp.HeaderAllow))
		})
	}
}

func (s *RequestMethodScenario) TestShouldRespondWithAppropriateResponseWithMethodHEAD() {
	testCases := []struct {
		name                  string
		uri                   string
		expectedStatus        int
		expectedContentLength bool
	}{
		{"RootPathShouldShowContentLengthAndRespondOK", AutheliaBaseURL, fasthttp.StatusOK, true},
		{"OpenAPISpecShouldShowContentLengthAndRespondOK", fmt.Sprintf("%s/api/openapi.yml", AutheliaBaseURL), fasthttp.StatusOK, true},
		{"LocalesShouldShowContentLengthAndRespondOK", fmt.Sprintf("%s/locales/en/portal.json", AutheliaBaseURL), fasthttp.StatusOK, true},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := http.NewRequest(fasthttp.MethodHead, tc.uri, nil)
			s.Assert().NoError(err)

			res, err := s.client.Do(req)

			s.Assert().NoError(err)
			s.Assert().Equal(tc.expectedStatus, res.StatusCode)

			if tc.expectedContentLength {
				s.Assert().NotEqual(0, res.ContentLength)
			} else {
				s.Assert().Equal(0, res.ContentLength)
			}

			data, err := io.ReadAll(res.Body)

			s.Assert().NoError(err)
			s.Assert().Len(data, 0)
		})
	}
}

func TestRunRequestMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewRequestMethodScenario())
}
