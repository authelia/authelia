package webauthn

import (
	"fmt"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func New(config *schema.Configuration) (provider *Provider, err error) {
	var (
		webauthnConfig *webauthn.Config
		webauthnImpl   *webauthn.WebAuthn
	)

	webauthnConfig = &webauthn.Config{
		RPDisplayName: config.Webauthn.DisplayName,
		RPID:          config.Server.Domain,
		RPOrigin:      fmt.Sprintf("https://%s%s", config.Server.Domain, config.Server.Path),
		Timeout:       config.Webauthn.Timeout,
		Debug:         config.Webauthn.Debug,
	}

	if config.Webauthn.AuthenticatorSelection != nil {
		switch config.Webauthn.AuthenticatorSelection.AuthenticatorAttachment {
		case protocol.Platform, protocol.CrossPlatform:
			webauthnConfig.AuthenticatorSelection.AuthenticatorAttachment = config.Webauthn.AuthenticatorSelection.AuthenticatorAttachment
		}

		switch config.Webauthn.AuthenticatorSelection.UserVerification {
		case protocol.VerificationRequired, protocol.VerificationPreferred, protocol.VerificationDiscouraged:
			webauthnConfig.AuthenticatorSelection.UserVerification = config.Webauthn.AuthenticatorSelection.UserVerification
		}

		webauthnConfig.AuthenticatorSelection.RequireResidentKey = &config.Webauthn.AuthenticatorSelection.RequireResidentKey
	}

	switch config.Webauthn.AttestationPreference {
	case protocol.PreferNoAttestation, protocol.PreferDirectAttestation, protocol.PreferIndirectAttestation:
		webauthnConfig.AttestationPreference = config.Webauthn.AttestationPreference
	default:
		webauthnConfig.AttestationPreference = protocol.PreferIndirectAttestation
	}

	if webauthnImpl, err = webauthn.New(webauthnConfig); err != nil {
		return nil, err
	}

	return &Provider{WebAuthn: webauthnImpl}, nil
}

type Provider struct {
	*webauthn.WebAuthn
}

func (p *Provider) NewUser(name string) (user *User) {
	return &User{
		ID:          1,
		Name:        name,
		DisplayName: name,
	}
}
