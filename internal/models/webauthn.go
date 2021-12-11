package models

import (
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

	for i, device := range w.Devices {
		aaguid, err := device.AAGUID.MarshalBinary()
		if err != nil {
			continue
		}

		credentials[i] = webauthn.Credential{
			ID:              device.KID,
			PublicKey:       device.PublicKey,
			AttestationType: device.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:    aaguid,
				SignCount: device.SignCount,
			},
		}
	}

	return credentials
}

// NewWebauthnDeviceFromCredential creates a WebauthnDevice from a webauthn.Credential.
func NewWebauthnDeviceFromCredential(username, description string, credential *webauthn.Credential) (device WebauthnDevice) {
	device.Username = username
	device.Description = description
	device.KID = credential.ID
	device.PublicKey = credential.PublicKey
	device.AttestationType = credential.AttestationType

	aaguid, _ := uuid.ParseBytes(credential.Authenticator.AAGUID)

	device.AAGUID = aaguid
	device.SignCount = credential.Authenticator.SignCount

	return device
}

// WebauthnDevice represents a Webauthn Device in the database storage.
type WebauthnDevice struct {
	ID              int       `db:"id"`
	Username        string    `db:"username"`
	Description     string    `db:"description"`
	KID             []byte    `db:"kid"`
	PublicKey       []byte    `db:"public_key"`
	AttestationType string    `db:"attestation_type"`
	Transports      []string  `db:"transports"`
	AAGUID          uuid.UUID `db:"aaguid"`
	SignCount       uint32    `db:"sign_count"`
}
