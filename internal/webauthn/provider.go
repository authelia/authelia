// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"fmt"
	"net/url"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewProvider returns a *Provider for the origin of the given context, resolving the relying party the origin belongs
// to when relying parties are configured.
func NewProvider(ctx Context) (provider *Provider, err error) {
	if ctx.GetConfiguration().WebAuthn.Disable {
		return nil, fmt.Errorf("webauthn is disabled")
	}

	var (
		origin *url.URL
	)

	if origin, err = ctx.GetOrigin(); err != nil {
		return nil, fmt.Errorf("error occurred determining the origin for the request: %w", err)
	}

	var opaqueOrigins []string

	rpid := origin.Hostname()
	origins := []string{origin.String()}
	base := ctx.GetConfiguration().WebAuthn.WebAuthnBase

	if len(ctx.GetConfiguration().WebAuthn.RelyingParties) != 0 {
		relyingPartyID, relyingParty := GetRelatedOriginConfigByOrigin(ctx.GetConfiguration().WebAuthn, origin)

		if relyingParty == nil {
			return nil, fmt.Errorf("error occurred finding the relying party: no related origin found for origin '%s'", origin.String())
		}

		rpid = relyingPartyID

		origins = relyingParty.StringOrigins()
		opaqueOrigins = relyingParty.OpaqueOrigins

		base = relyingParty.WebAuthnBase
	}

	config := &gowebauthn.Config{
		RPID:                  rpid,
		RPDisplayName:         base.DisplayName,
		RPOrigins:             origins,
		RPOpaqueOrigins:       opaqueOrigins,
		AttestationPreference: base.ConveyancePreference,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			AuthenticatorAttachment: base.SelectionCriteria.Attachment,
			ResidentKey:             base.SelectionCriteria.Discoverability,
			UserVerification:        base.SelectionCriteria.UserVerification,
		},
		Debug:                false,
		EncodeUserIDAsString: false,
		Timeouts: gowebauthn.TimeoutsConfig{
			Login: gowebauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    base.Timeout,
				TimeoutUVD: base.Timeout,
			},
			Registration: gowebauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    base.Timeout,
				TimeoutUVD: base.Timeout,
			},
		},
		MDS:                               ctx.GetWebAuthnMetaDataProvider(),
		ExtensionsUnsolicitedOutputPolicy: protocol.UnsolicitedOutputPolicyReject,
	}

	switch base.SelectionCriteria.Attachment {
	case protocol.Platform, protocol.CrossPlatform:
		config.AuthenticatorSelection.AuthenticatorAttachment = base.SelectionCriteria.Attachment
	}

	switch base.SelectionCriteria.Discoverability {
	case protocol.ResidentKeyRequirementRequired:
		config.AuthenticatorSelection.RequireResidentKey = protocol.ResidentKeyRequired()
	case protocol.ResidentKeyRequirementPreferred, protocol.ResidentKeyRequirementDiscouraged:
		config.AuthenticatorSelection.RequireResidentKey = protocol.ResidentKeyNotRequired()
	}

	webauthn, err := gowebauthn.New(config)
	if err != nil {
		return nil, err
	}

	return &Provider{
		WebAuthn: webauthn,
		Config:   NewProviderConfig(ctx.GetConfiguration().WebAuthn, base),
	}, nil
}

// Provider is a *gowebauthn.WebAuthn along with the Authelia configuration which applies to the relying party it was
// built for.
type Provider struct {
	*gowebauthn.WebAuthn

	Config ProviderConfig
}

// NewProviderConfig returns the ProviderConfig which combines the global WebAuthn options with the options of the
// relying party the provider was built for.
func NewProviderConfig(wconfig schema.WebAuthn, base schema.WebAuthnBase) (config ProviderConfig) {
	config = ProviderConfig{
		Disable:              wconfig.Disable,
		EnablePasskeyLogin:   wconfig.EnablePasskeyLogin,
		EnablePasskey2FA:     wconfig.EnablePasskey2FA,
		EnablePasskeyUpgrade: wconfig.EnablePasskeyUpgrade,
		WebAuthnBase:         base,
	}

	return config
}

// ProviderConfig is the Authelia configuration which applies to a Provider.
type ProviderConfig struct {
	Disable              bool
	EnablePasskeyLogin   bool
	EnablePasskey2FA     bool
	EnablePasskeyUpgrade bool

	schema.WebAuthnBase
}
