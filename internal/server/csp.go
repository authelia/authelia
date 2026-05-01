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

	raw := ctx.RequestCtx.QueryArgs().Peek(oidc.FormParameterFlowID)
	if len(raw) == 0 {
		return ""
	}

	flowID, err := uuid.ParseBytes(raw)
	if err != nil {
		return ""
	}

	if ctx.Providers.StorageProvider == nil {
		return ""
	}

	consent, err := ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, flowID)
	if err != nil || consent == nil {
		return ""
	}

	if ctx.Providers.OpenIDConnect == nil {
		return ""
	}

	client, err := ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID)
	if err != nil || client == nil {
		return ""
	}

	registered, ok := client.(*oidc.RegisteredClient)
	if !ok {
		return ""
	}

	logo := registered.GetLogoURI()
	if logo == nil || logo.Scheme != "https" || logo.Host == "" {
		return ""
	}

	return " https://" + logo.Host
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
