package suites

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/oidc"
)

// WARNING: This scenario is intended to be used with TLS enabled in the authelia backend.

type OIDCClientCredentialsScenario struct {
	suite.Suite

	client   *http.Client
	metadata map[string]any
}

func NewOIDCClientCredentialsScenario() *OIDCClientCredentialsScenario {
	return &OIDCClientCredentialsScenario{}
}

func (s *OIDCClientCredentialsScenario) SetupSuite() {
	s.client = NewHTTPClient()

	metadata, err := s.discoverMetadata(60 * time.Second)
	require.NoError(s.T(), err)

	s.metadata = metadata
}

func (s *OIDCClientCredentialsScenario) discoverMetadata(timeout time.Duration) (metadata map[string]any, err error) {
	endpoint := fmt.Sprintf("%s/.well-known/openid-configuration", LoginBaseURL)

	deadline := time.Now().Add(timeout)

	for {
		if metadata, err = s.fetchMetadata(endpoint); err == nil {
			return metadata, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("failed to discover OpenID Connect metadata within %s: %w", timeout, err)
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *OIDCClientCredentialsScenario) fetchMetadata(endpoint string) (metadata map[string]any, err error) {
	resp, err := s.client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(body, &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (s *OIDCClientCredentialsScenario) strDiscovery(name string) string {
	value, ok := s.metadata[name].(string)
	require.Truef(s.T(), ok, "metadata does not contain a valid '%s' value", name)

	return value
}

func (s *OIDCClientCredentialsScenario) TestShouldIssueClientCredentialsBearerToken() {
	var (
		resp *http.Response
		body []byte
		err  error
	)

	issuer := s.strDiscovery("issuer")

	clientID := "client-credentials-opaque"
	clientSecret := "foobar"

	tokenData := url.Values{}
	tokenData.Set("grant_type", "client_credentials")
	tokenData.Set("client_id", clientID)
	tokenData.Set("client_secret", clientSecret)

	resp, err = s.client.PostForm(s.strDiscovery("token_endpoint"), tokenData)
	require.NoError(s.T(), err)

	body, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(s.T(), err)

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var tokenResponse map[string]any

	require.NoError(s.T(), json.Unmarshal(body, &tokenResponse))

	assert.Equal(s.T(), "bearer", tokenResponse["token_type"])

	accessToken, ok := tokenResponse["access_token"].(string)
	require.True(s.T(), ok)

	assert.True(s.T(), strings.HasPrefix(accessToken, "authelia_at_"))

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("token", accessToken)

	resp, err = s.client.PostForm(s.strDiscovery("introspection_endpoint"), data)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.NoError(s.T(), err)

	var introspection map[string]any

	assert.NoError(s.T(), json.Unmarshal(body, &introspection))

	s.verifyMandatedClaims(s.T(), "introspection response", introspection, []string{
		oidc.ClaimActive, oidc.ClaimClientIdentifier, oidc.ClaimExpirationTime, oidc.ClaimIssuedAt,
	})

	assert.Equal(s.T(), true, introspection[oidc.ClaimActive])
	assert.Equal(s.T(), clientID, introspection[oidc.ClaimClientIdentifier])
	assert.Equal(s.T(), clientID, introspection[oidc.ClaimSubject])
	assert.Equal(s.T(), issuer, introspection[oidc.ClaimIssuer])
}

func (s *OIDCClientCredentialsScenario) TestShouldIssueClientCredentialsJWTAccessToken() {
	var (
		resp *http.Response
		body []byte
		err  error
	)

	issuer := s.strDiscovery("issuer")
	clientID := "client-credentials-jwt"
	clientSecret := "foobar"

	tokenData := url.Values{}
	tokenData.Set("grant_type", "client_credentials")
	tokenData.Set("client_id", clientID)
	tokenData.Set("client_secret", clientSecret)

	resp, err = s.client.PostForm(s.strDiscovery("token_endpoint"), tokenData)
	require.NoError(s.T(), err)

	body, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(s.T(), err)

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var tokenResponse map[string]any

	require.NoError(s.T(), json.Unmarshal(body, &tokenResponse))

	assert.Equal(s.T(), "bearer", tokenResponse["token_type"])

	accessToken, ok := tokenResponse["access_token"].(string)
	require.True(s.T(), ok)

	assert.False(s.T(), strings.HasPrefix(accessToken, "authelia_at_"))

	parts := strings.Split(accessToken, ".")
	require.Len(s.T(), parts, 3)

	rawHeader, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(s.T(), err)

	var header map[string]any

	require.NoError(s.T(), json.Unmarshal(rawHeader, &header))

	assert.Equal(s.T(), "at+jwt", header["typ"])
	assert.Equal(s.T(), "RS256", header["alg"])

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(s.T(), err)

	var claims map[string]any

	require.NoError(s.T(), json.Unmarshal(payload, &claims))

	s.verifyMandatedClaims(s.T(), "access token", claims, []string{
		oidc.ClaimIssuer, oidc.ClaimSubject, oidc.ClaimAudience, oidc.ClaimExpirationTime,
		oidc.ClaimIssuedAt, oidc.ClaimNotBefore, oidc.ClaimJWTID, oidc.ClaimClientIdentifier,
		oidc.ClaimScopeNonStandard,
	})

	assert.Equal(s.T(), clientID, claims[oidc.ClaimClientIdentifier])
	assert.Equal(s.T(), clientID, claims[oidc.ClaimSubject])
	assert.Equal(s.T(), issuer, claims[oidc.ClaimIssuer])

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("token", accessToken)

	resp, err = s.client.PostForm(s.strDiscovery("introspection_endpoint"), data)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.NoError(s.T(), err)

	var introspection map[string]any

	assert.NoError(s.T(), json.Unmarshal(body, &introspection))

	s.verifyMandatedClaims(s.T(), "introspection response", introspection, []string{
		oidc.ClaimActive, oidc.ClaimClientIdentifier, oidc.ClaimExpirationTime, oidc.ClaimIssuedAt,
	})

	assert.Equal(s.T(), true, introspection[oidc.ClaimActive])
	assert.Equal(s.T(), clientID, introspection[oidc.ClaimClientIdentifier])
	assert.Equal(s.T(), clientID, introspection[oidc.ClaimSubject])
	assert.Equal(s.T(), issuer, introspection[oidc.ClaimIssuer])
}

func (s *OIDCClientCredentialsScenario) verifyMandatedClaims(t *testing.T, location string, claims map[string]any, mandated []string) {
	for _, claim := range mandated {
		t.Run(fmt.Sprintf("%s/%s", location, claim), func(t *testing.T) {
			value, ok := claims[claim]
			require.Truef(t, ok, "claim '%s' is mandated for the client credentials flow in the %s but was not present", claim, location)
			assert.NotNilf(t, value, "claim '%s' is mandated for the client credentials flow in the %s but was null", claim, location)
		})
	}
}

func TestRunOIDCClientCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOIDCClientCredentialsScenario())
}
