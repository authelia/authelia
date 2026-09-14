// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"strings"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func resolveOIDCConsentLogoURI(ctx *middlewares.AutheliaCtx) string {
	if !isOIDCConsentShellPath(string(ctx.Path())) {
		return ""
	}

	if ctx.Providers.StorageProvider == nil || ctx.Providers.OpenIDConnect == nil {
		return ""
	}

	clientID := resolveOIDCConsentClientID(ctx)
	if clientID == "" {
		return ""
	}

	client, err := ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID)
	if err != nil || client == nil {
		return ""
	}

	logo := client.GetLogoURI()
	if logo == nil || logo.Scheme != "https" || logo.Host == "" {
		return ""
	}

	return " https://" + logo.Host
}

func resolveOIDCConsentClientID(ctx *middlewares.AutheliaCtx) string {
	args := ctx.QueryArgs()

	if raw := args.Peek(oidc.FormParameterFlowID); len(raw) != 0 {
		flowID, err := uuid.ParseBytes(raw)
		if err != nil {
			return ""
		}

		consent, err := ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, flowID)
		if err != nil || consent == nil {
			return ""
		}

		return consent.ClientID
	}

	if raw := args.Peek(oidc.FormParameterUserCode); len(raw) != 0 {
		signature, err := ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(ctx, string(raw))
		if err != nil {
			return ""
		}

		device, err := ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, signature)
		if err != nil || device == nil {
			return ""
		}

		return device.ClientID
	}

	return ""
}

func isOIDCConsentShellPath(path string) bool {
	return path == oidc.FrontendEndpointPathConsentDecision ||
		path == oidc.FrontendEndpointPathConsentDeviceAuthorization
}

func expandCSPTemplate(tmpl, nonce, oidcClientLogoURIs string) string {
	out := strings.ReplaceAll(tmpl, placeholderCSPNonce, nonce)
	out = strings.ReplaceAll(out, placeholderCSPOIDCClientLogoURIs, oidcClientLogoURIs)

	return out
}
