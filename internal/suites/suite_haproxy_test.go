// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type HAProxySuite struct {
	*RodSuite
}

func NewHAProxySuite() *HAProxySuite {
	return &HAProxySuite{
		RodSuite: NewRodSuite(haproxySuiteName),
	}
}

func (s *HAProxySuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *HAProxySuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *HAProxySuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *HAProxySuite) TestShouldNotTrustForgedForwardedHeaders() {
	client := NewHTTPClient()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/headers", SecureBaseURL), nil)
	s.Require().NoError(err)

	req.Header.Set(fasthttp.HeaderAccept, "application/json")
	req.Header.Set(fasthttp.HeaderXForwardedFor, "192.0.2.1")

	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().NoError(resp.Body.Close())

	s.Equal(http.StatusUnauthorized, resp.StatusCode)

	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/headers", PublicBaseURL), nil)
	s.Require().NoError(err)

	req.Header.Set(fasthttp.HeaderXForwardedFor, "192.0.2.1")
	req.Header.Set(fasthttp.HeaderXForwardedProto, "http")

	resp, err = client.Do(req)
	s.Require().NoError(err)

	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	body := struct {
		Headers map[string][]string `json:"headers"`
	}{}

	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))

	s.Require().Len(body.Headers[fasthttp.HeaderXForwardedFor], 1)
	s.NotContains(body.Headers[fasthttp.HeaderXForwardedFor][0], "192.0.2.1")
	s.Equal([]string{"https"}, body.Headers[fasthttp.HeaderXForwardedProto])
}

func (s *HAProxySuite) TestShouldNotForwardForgedRemoteHeaders() {
	client := NewHTTPClient()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/headers", PublicBaseURL), nil)
	s.Require().NoError(err)

	req.Header.Set("Remote-User", "admin")
	req.Header.Set("Remote-Groups", "admins")
	req.Header.Set("Remote-Name", "Admin")
	req.Header.Set("Remote-Email", "admin@example.com")

	resp, err := client.Do(req)
	s.Require().NoError(err)

	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	body := struct {
		Headers map[string][]string `json:"headers"`
	}{}

	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))

	for _, header := range []string{"Remote-User", "Remote-Groups", "Remote-Name", "Remote-Email"} {
		s.NotContains(body.Headers, header)
	}
}

func (s *HAProxySuite) TestShouldForwardAuthzDenialResponses() {
	client := NewHTTPClient()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/secret.html", SecureBaseURL), nil)
	s.Require().NoError(err)

	req.Header.Set(fasthttp.HeaderAccept, "application/json")

	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().NoError(resp.Body.Close())

	s.Equal(http.StatusUnauthorized, resp.StatusCode)
	s.Contains(resp.Header.Get(fasthttp.HeaderLocation), LoginBaseURL)

	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/secret.html", SecureBaseURL), nil)
	s.Require().NoError(err)

	req.Header.Set(fasthttp.HeaderAccept, "text/html")

	resp, err = client.Do(req)
	s.Require().NoError(err)
	s.Require().NoError(resp.Body.Close())

	s.Equal(http.StatusFound, resp.StatusCode)
	s.Contains(resp.Header.Get(fasthttp.HeaderLocation), LoginBaseURL)
}

func TestHAProxySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewHAProxySuite())
}
