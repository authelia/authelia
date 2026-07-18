package webauthn

import (
	"fmt"
	"net/url"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

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

	rpid := origin.Hostname()
	origins := []string{origin.String()}
	base := ctx.GetConfiguration().WebAuthn.WebAuthnBase

	if len(ctx.GetConfiguration().WebAuthn.RelyingParties) != 0 {
		relyingPartyID, relyingParty := GetRelatedOriginConfigByOrigin(ctx.GetConfiguration().WebAuthn, origin)

		if relyingParty == nil {
			return nil, fmt.Errorf("error occurred finding the relying party: no related origin found for origin '%s'", origin.String())
		}

		rpid = relyingPartyID

		origins = append(relyingParty.StringOrigins(), relyingParty.OpaqueOrigins...)

		base = relyingParty.WebAuthnBase
	}

	config := &gowebauthn.Config{
		RPID:                  rpid,
		RPDisplayName:         base.DisplayName,
		RPOrigins:             origins,
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
		MDS: ctx.GetWebAuthnMetaDataProvider(),
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

type Provider struct {
	*gowebauthn.WebAuthn

	Config ProviderConfig
}

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

type ProviderConfig struct {
	Disable              bool
	EnablePasskeyLogin   bool
	EnablePasskey2FA     bool
	EnablePasskeyUpgrade bool

	schema.WebAuthnBase
}
