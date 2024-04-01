package oidc_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewClaimRequests(t *testing.T) {
	form := url.Values{}

	form.Set(oidc.FormParameterClaims, `{"id_token":{"sub":{"value":"aaaa"}}}`)

	requests, err := oidc.NewClaimRequests(form)
	require.NoError(t, err)

	assert.NotNil(t, requests)

	var (
		requested string
		ok        bool
	)

	requested, ok = requests.MatchesSubject("aaaa")
	assert.Equal(t, "aaaa", requested)
	assert.True(t, ok)

	requested, ok = requests.MatchesSubject("aaaaa")
	assert.Equal(t, "aaaa", requested)
	assert.False(t, ok)
}
