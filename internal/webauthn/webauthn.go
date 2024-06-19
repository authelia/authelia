package webauthn

import (
	"fmt"

	"github.com/go-webauthn/webauthn/metadata"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

func New(config *schema.Configuration) (provider *ProductionProvider, err error) {
	if config == nil || config.WebAuthn.Disable {
		return nil, nil
	}

	var (
		w   *webauthn.WebAuthn
		mds metadata.Provider
	)

	if mds, err = NewMetaDataProvider(config); err != nil {
		return nil, err
	}

	wc := &webauthn.Config{
		RPDisplayName:         config.WebAuthn.DisplayName,
		AttestationPreference: config.WebAuthn.ConveyancePreference,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			AuthenticatorAttachment: config.WebAuthn.SelectionCriteria.Attachment,
			ResidentKey:             config.WebAuthn.SelectionCriteria.Discoverability,
			UserVerification:        config.WebAuthn.SelectionCriteria.UserVerification,
		},
		Debug:                false,
		EncodeUserIDAsString: true,
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    config.WebAuthn.Timeout,
				TimeoutUVD: config.WebAuthn.Timeout,
			},
			Registration: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    config.WebAuthn.Timeout,
				TimeoutUVD: config.WebAuthn.Timeout,
			},
		},
		MDS: mds,
	}

	switch config.WebAuthn.SelectionCriteria.Attachment {
	case protocol.Platform, protocol.CrossPlatform:
		wc.AuthenticatorSelection.AuthenticatorAttachment = config.WebAuthn.SelectionCriteria.Attachment
	}

	switch config.WebAuthn.SelectionCriteria.Discoverability {
	case protocol.ResidentKeyRequirementRequired:
		wc.AuthenticatorSelection.RequireResidentKey = protocol.ResidentKeyRequired()
	default:
		wc.AuthenticatorSelection.RequireResidentKey = protocol.ResidentKeyNotRequired()
	}

	if w, err = webauthn.New(wc); err != nil {
		return nil, err
	}

	return &ProductionProvider{
		config:   config,
		metadata: mds,
		webauthn: w,
	}, nil
}

type ProductionProvider struct {
	config   *schema.Configuration
	metadata metadata.Provider
	webauthn *webauthn.WebAuthn
}

func (p *ProductionProvider) ValidateCredentialAllowed(credential *model.WebAuthnCredential) (err error) {
	if len(p.config.WebAuthn.Filtering.PermittedAAGUIDs) != 0 {
		for _, aaguid := range p.config.WebAuthn.Filtering.PermittedAAGUIDs {
			if credential.AAGUID.UUID == aaguid {
				return nil
			}
		}
		return fmt.Errorf("error checking webauthn AAGUID: filters have been configured which explicitly require only permitted AAGUID's be used and '%s' is not permitted", credential.AAGUID.UUID)
	}

	for _, aaguid := range p.config.WebAuthn.Filtering.ProhibitedAAGUIDs {
		if credential.AAGUID.UUID == aaguid {
			return fmt.Errorf("error checking webauthn AAGUID: filters have been configured which prohibit the AAGUID '%s' from registration", aaguid)
		}
	}

	return nil
}
