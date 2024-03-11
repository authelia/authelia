package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/mocks"
)

type passwordPolicyResponseBody struct {
	Status string
	Data   PasswordPolicyBody
}

type PasswordPolicySuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *PasswordPolicySuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
}

func (s *PasswordPolicySuite) TearDownTest() {
	s.mock.Close()
}

func (s *PasswordPolicySuite) TestShouldBeDisabled() {
	s.mock.Ctx.Configuration.PasswordPolicy.ZXCVBN.Enabled = false
	s.mock.Ctx.Configuration.PasswordPolicy.Standard.Enabled = false

	PasswordPolicyConfigurationGET(s.mock.Ctx)

	response := &passwordPolicyResponseBody{}
	err := json.Unmarshal(s.mock.Ctx.Response.Body(), response)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "disabled", response.Data.Mode)
}

func (s *PasswordPolicySuite) TestShouldBeStandard() {
	s.mock.Ctx.Configuration.PasswordPolicy.ZXCVBN.Enabled = false
	s.mock.Ctx.Configuration.PasswordPolicy.Standard.Enabled = true
	s.mock.Ctx.Configuration.PasswordPolicy.Standard.MinLength = 4
	s.mock.Ctx.Configuration.PasswordPolicy.Standard.MaxLength = 8

	PasswordPolicyConfigurationGET(s.mock.Ctx)

	response := &passwordPolicyResponseBody{}
	err := json.Unmarshal(s.mock.Ctx.Response.Body(), response)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "standard", response.Data.Mode)
	assert.Equal(s.T(), 4, response.Data.MinLength)
	assert.Equal(s.T(), 8, response.Data.MaxLength)
}

func (s *PasswordPolicySuite) TestShouldBeZXCVBN() {
	s.mock.Ctx.Configuration.PasswordPolicy.ZXCVBN.Enabled = true
	s.mock.Ctx.Configuration.PasswordPolicy.Standard.Enabled = false

	PasswordPolicyConfigurationGET(s.mock.Ctx)

	response := &passwordPolicyResponseBody{}
	err := json.Unmarshal(s.mock.Ctx.Response.Body(), response)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "zxcvbn", response.Data.Mode)
}

func TestRunPasswordPolicySuite(t *testing.T) {
	s := new(PasswordPolicySuite)
	suite.Run(t, s)
}
