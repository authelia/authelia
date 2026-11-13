// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"fmt"
	"net/http"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/templates"
)

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(config *schema.Configuration, store storage.Provider, templates *templates.Provider) (provider *OpenIDConnectProvider, err error) {
	if config == nil || config.IdentityProviders.OIDC == nil {
		return nil, nil
	}

	var idjag *IDJAGConfig

	if idjag, err = NewIDJAGConfig(config.IdentityProviders.OIDC); err != nil {
		return nil, fmt.Errorf("error occurred initializing the identity assertion authorization grant: %w", err)
	}

	issuer := NewIssuer(config.IdentityProviders.OIDC.JSONWebKeys)

	provider = &OpenIDConnectProvider{
		Issuer: issuer,
		Config: NewConfig(config.IdentityProviders.OIDC, issuer, templates),
	}

	provider.IDJAG = idjag
	provider.Store = NewStore(config, store, provider.Config)

	provider.Provider = oauthelia2.New(provider.Store, provider.Config)

	provider.LoadHandlers(provider.Store)

	return provider, nil
}

// WriteDynamicAuthorizeError writes the authorization error appropriate for the given requester.
func (p *OpenIDConnectProvider) WriteDynamicAuthorizeError(ctx Context, rw http.ResponseWriter, requester oauthelia2.Requester, err error) {
	switch r := requester.(type) {
	case oauthelia2.DeviceAuthorizeRequester:
		p.WriteRFC8628UserAuthorizeError(ctx, rw, r, err)
	case oauthelia2.AuthorizeRequester:
		p.WriteAuthorizeError(ctx, rw, r, err)
	}
}
