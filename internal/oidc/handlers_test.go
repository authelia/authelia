package oidc_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestAuthorizationServerIssuerIdentificationHandler_HandleAuthorizeEndpointRequest(t *testing.T) {
	testCases := []struct {
		name     string
		issuer   string
		have     fosite.AuthorizeRequester
		expected url.Values
	}{
		{
			"ShouldAddIssuer",
			"https://auth.example.com",
			&fosite.AuthorizeRequest{},
			url.Values{"iss": []string{"https://auth.example.com"}},
		},
		{
			"ShouldAddIssuerResponseModeFormPost",
			"https://auth.example.com",
			&fosite.AuthorizeRequest{ResponseMode: fosite.ResponseModeFormPost},
			url.Values{"iss": []string{"https://auth.example.com"}},
		},
		{
			"ShouldAddIssuerResponseModeQuery",
			"https://auth.example.com",
			&fosite.AuthorizeRequest{ResponseMode: fosite.ResponseModeQuery},
			url.Values{"iss": []string{"https://auth.example.com"}},
		},
		{
			"ShouldAddIssuerResponseModeFragment",
			"https://auth.example.com",
			&fosite.AuthorizeRequest{ResponseMode: fosite.ResponseModeFragment},
			url.Values{"iss": []string{"https://auth.example.com"}},
		},
		{
			"ShouldAddIssuerResponseModeDefault",
			"https://auth.example.com",
			&fosite.AuthorizeRequest{ResponseMode: fosite.ResponseModeDefault},
			url.Values{"iss": []string{"https://auth.example.com"}},
		},
		{
			"ShouldNotAddIssuerResponseModeFormPostJWT",
			"https://auth.example.com",
			&fosite.AuthorizeRequest{ResponseMode: oidc.ResponseModeFormPostJWT},
			url.Values{},
		},
		{
			"ShouldNotAddIssuerResponseModeQueryJWT",
			"https://auth.example.com",
			&fosite.AuthorizeRequest{ResponseMode: oidc.ResponseModeQueryJWT},
			url.Values{},
		},
		{
			"ShouldNotAddIssuerResponseModeFragmentJWT",
			"https://auth.example.com",
			&fosite.AuthorizeRequest{ResponseMode: oidc.ResponseModeFragmentJWT},
			url.Values{},
		},
		{
			"ShouldNotAddIssuerResponseModeJWT",
			"https://auth.example.com",
			&fosite.AuthorizeRequest{ResponseMode: oidc.ResponseModeJWT},
			url.Values{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{Issuers: oidc.IssuersConfig{
				AuthorizationServerIssuerIdentification: tc.issuer,
			}}

			handler := &oidc.AuthorizationServerIssuerIdentificationHandler{Config: config}

			responder := fosite.NewAuthorizeResponse()

			ctx := context.TODO()

			assert.NoError(t, handler.HandleAuthorizeEndpointRequest(ctx, tc.have, responder))

			assert.Equal(t, tc.expected, responder.GetParameters())
		})
	}
}
