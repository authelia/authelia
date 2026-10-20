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

type EnvoySuite struct {
	*RodSuite
}

func NewEnvoySuite() *EnvoySuite {
	return &EnvoySuite{
		RodSuite: NewRodSuite(envoySuiteName),
	}
}

func (s *EnvoySuite) Test1FAScenario() {
	suite.Run(s.T(), New1FAScenario())
}

func (s *EnvoySuite) Test2FATOTPScenario() {
	suite.Run(s.T(), New2FATOTPScenario())
}

func (s *EnvoySuite) TestCustomHeaders() {
	suite.Run(s.T(), NewCustomHeadersScenario())
}

func (s *EnvoySuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *EnvoySuite) TestChangePasswordScenario() {
	suite.Run(s.T(), NewChangePasswordScenario())
}

func (s *EnvoySuite) TestShouldNotTrustForgedForwardedHeaders() {
	client := NewHTTPClient()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/headers", SecureBaseURL), nil)
	s.Require().NoError(err)

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

func TestEnvoySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewEnvoySuite())
}
