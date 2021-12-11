package models

import (
	"encoding/hex"
	"fmt"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/google/uuid"
)

// WebauthnUser is an object to represent a user for the Webauthn lib.
type WebauthnUser struct {
	Username    string
	DisplayName string
	Devices     []WebauthnDevice
}

// HasFIDOU2F returns true if the user has any attestation type `fido-u2f` devices.
func (w WebauthnUser) HasFIDOU2F() bool {
	for _, c := range w.Devices {
		if c.AttestationType == "fido-u2f" {
			return true
		}
	}

	return false
}

// WebAuthnID implements the webauthn.User interface.
func (w WebauthnUser) WebAuthnID() []byte {
	return []byte(w.Username)
}

// WebAuthnName implements the webauthn.User  interface.
func (w WebauthnUser) WebAuthnName() string {
	return w.Username
}

// WebAuthnDisplayName implements the webauthn.User interface.
func (w WebauthnUser) WebAuthnDisplayName() string {
	return w.DisplayName
}

// WebAuthnIcon implements the webauthn.User interface.
func (w WebauthnUser) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials implements the webauthn.User interface.
func (w WebauthnUser) WebAuthnCredentials() (credentials []webauthn.Credential) {
	credentials = make([]webauthn.Credential, len(w.Devices))

	fmt.Printf("devices: %d\n", len(w.Devices))

	for i, device := range w.Devices {
		aaguid, err := device.AAGUID.MarshalBinary()
		if err != nil {
			continue
		}

		credentials[i] = webauthn.Credential{
			ID:              device.KID.Bytes(),
			PublicKey:       device.PublicKey,
			AttestationType: device.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:       aaguid,
				SignCount:    device.SignCount,
				CloneWarning: device.CloneWarning,
			},
		}

		fmt.Printf("decoded device - id: %x, attestation: %s, aaguid: %x, sign count: %d\n", credentials[i].ID, credentials[i].AttestationType, credentials[i].Authenticator.AAGUID, credentials[i].Authenticator.SignCount)
	}

	return credentials
}

// WebAuthnCredentialDescriptors decodes the users credentials into protocol.CredentialDescriptor's.
func (w WebauthnUser) WebAuthnCredentialDescriptors() (descriptors []protocol.CredentialDescriptor) {
	descriptors = make([]protocol.CredentialDescriptor, len(w.Devices))

	fmt.Printf("descriptors: %d\n", len(w.Devices))

	for i, device := range w.Devices {
		descriptor := protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: device.KID.Bytes(),
		}

		for _, t := range device.Transport {
			transport := protocol.AuthenticatorTransport(t)

			switch transport {
			case protocol.Internal, protocol.USB, protocol.NFC, protocol.BLE:
				descriptor.Transport = append(descriptor.Transport, transport)
			}
		}

		descriptors[i] = descriptor

		fmt.Printf("decoded descriptor - id: %x, type: %s, transport: %+v\n", descriptors[i].CredentialID, descriptors[i].Type, descriptors[i].Transport)
	}

	return descriptors
}

// NewWebauthnDeviceFromCredential creates a WebauthnDevice from a webauthn.Credential.
func NewWebauthnDeviceFromCredential(username, description string, credential *webauthn.Credential) (device WebauthnDevice) {
	device = WebauthnDevice{
		Username:        username,
		Description:     description,
		KID:             Hexadecimal{value: credential.ID},
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		SignCount:       credential.Authenticator.SignCount,
		CloneWarning:    credential.Authenticator.CloneWarning,
	}

	device.AAGUID, _ = uuid.Parse(hex.EncodeToString(credential.Authenticator.AAGUID))

	return device
}

// WebauthnDevice represents a Webauthn Device in the database storage.
type WebauthnDevice struct {
	ID              int         `db:"id"`
	Username        string      `db:"username"`
	Description     string      `db:"description"`
	KID             Hexadecimal `db:"kid"`
	PublicKey       []byte      `db:"public_key"`
	AttestationType string      `db:"attestation_type"`
	Transport       []string    `db:"transport"`
	AAGUID          uuid.UUID   `db:"aaguid"`
	SignCount       uint32      `db:"sign_count"`
	CloneWarning    bool        `db:"clone_warning"`
}
