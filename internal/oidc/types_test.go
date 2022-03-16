package oidc

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	formValues.Set("nonce", "abc123xyzauthelia")

	request := &fosite.AuthorizeRequest{
		Request: fosite.Request{
			ID:     requestID.String(),
			Form:   formValues,
			Client: &InternalClient{ID: "example"},
		},
	}

	extra := map[string]interface{}{
		"preferred_username": "john",
	}

	requested := time.Unix(1647332518, 0)
	authAt := time.Unix(1647332500, 0)
	issuer := "https://example.com"

	session := NewSessionWithAuthorizeRequest(issuer, "primary", subject.String(), "john", extra, authAt, requested, request)

	require.NotNil(t, session)
	require.NotNil(t, session.Extra)
	require.NotNil(t, session.Headers)
	require.NotNil(t, session.Headers.Extra)
	require.NotNil(t, session.Claims)
	require.NotNil(t, session.Claims.Extra)

	assert.Equal(t, "abc123xyzauthelia", session.Claims.Nonce)
	assert.Equal(t, subject.String(), session.Claims.Subject)
	assert.Equal(t, subject.String(), session.Subject)
	assert.Equal(t, issuer, session.Claims.Issuer)
	assert.Equal(t, "primary", session.Headers.Get("kid"))
	assert.Equal(t, "example", session.ClientID)
	assert.Equal(t, requested, session.Claims.RequestedAt)
	assert.Equal(t, authAt, session.Claims.AuthTime)
	assert.Greater(t, session.Claims.IssuedAt.Unix(), authAt.Unix())
	assert.Equal(t, "john", session.Username)

	require.Contains(t, session.Claims.Extra, "preferred_username")
	assert.Equal(t, "john", session.Claims.Extra["preferred_username"])

	session = NewSessionWithAuthorizeRequest(issuer, "primary", subject.String(), "john", nil, authAt, requested, request)

	require.NotNil(t, session)
	require.NotNil(t, session.Claims)
	assert.NotNil(t, session.Claims.Extra)
}
