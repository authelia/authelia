package oidc

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestNewSession(t *testing.T) {
	session := NewSession()

	require.NotNil(t, session)

	assert.Equal(t, "", session.ClientID)
	assert.Equal(t, "", session.Username)
	assert.Equal(t, "", session.Subject)
	require.NotNil(t, session.Claims)
	assert.NotNil(t, session.Claims.Extra)
	assert.NotNil(t, session.Extra)
	require.NotNil(t, session.Headers)
	assert.NotNil(t, session.Headers.Extra)
}

func TestNewSessionWithAuthorizeRequest(t *testing.T) {
	requestID := uuid.New()
	subject := uuid.New()

	formValues := url.Values{}

	formValues.Set(ClaimNonce, "abc123xyzauthelia")

	request := &fosite.AuthorizeRequest{
		Request: fosite.Request{
			ID:     requestID.String(),
			Form:   formValues,
			Client: &BaseClient{ID: "example"},
		},
	}

	extra := map[string]any{
		ClaimPreferredUsername: "john",
	}

	requested := time.Unix(1647332518, 0)
	authAt := time.Unix(1647332500, 0)
	issuer := "https://example.com"
	amr := []string{AMRPasswordBasedAuthentication}

	consent := &model.OAuth2ConsentSession{
		ChallengeID: uuid.New(),
		RequestedAt: requested,
		Subject:     uuid.NullUUID{UUID: subject, Valid: true},
	}

	session := NewSessionWithAuthorizeRequest(MustParseRequestURI(issuer), "primary", "john", amr, extra, authAt, consent, request)

	require.NotNil(t, session)
	require.NotNil(t, session.Extra)
	require.NotNil(t, session.Headers)
	require.NotNil(t, session.Headers.Extra)
	require.NotNil(t, session.Claims)
	require.NotNil(t, session.Claims.Extra)
	require.NotNil(t, session.Claims.AuthenticationMethodsReferences)

	assert.Equal(t, subject.String(), session.Subject)
	assert.Equal(t, "example", session.ClientID)
	assert.Greater(t, session.Claims.IssuedAt.Unix(), authAt.Unix())
	assert.Equal(t, "john", session.Username)

	assert.Equal(t, "abc123xyzauthelia", session.Claims.Nonce)
	assert.Equal(t, subject.String(), session.Claims.Subject)
	assert.Equal(t, amr, session.Claims.AuthenticationMethodsReferences)
	assert.Equal(t, authAt, session.Claims.AuthTime)
	assert.Equal(t, requested, session.Claims.RequestedAt)
	assert.Equal(t, issuer, session.Claims.Issuer)
	assert.Equal(t, "john", session.Claims.Extra[ClaimPreferredUsername])

	assert.Equal(t, "primary", session.Headers.Get(JWTHeaderKeyIdentifier))

	consent = &model.OAuth2ConsentSession{
		ChallengeID: uuid.New(),
		RequestedAt: requested,
	}

	session = NewSessionWithAuthorizeRequest(MustParseRequestURI(issuer), "primary", "john", nil, nil, authAt, consent, request)

	require.NotNil(t, session)
	require.NotNil(t, session.Claims)
	assert.NotNil(t, session.Claims.Extra)
	assert.Nil(t, session.Claims.AuthenticationMethodsReferences)
}

func MustParseRequestURI(input string) *url.URL {
	if requestURI, err := url.ParseRequestURI(input); err != nil {
		panic(err)
	} else {
		return requestURI
	}
}
