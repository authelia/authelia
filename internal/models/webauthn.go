package models

import (
	"encoding/hex"
	"net"
	"strings"
	"time"

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

	var credential webauthn.Credential

	for i, device := range w.Devices {
		aaguid, err := device.AAGUID.MarshalBinary()
		if err != nil {
			continue
		}

		credential = webauthn.Credential{
			ID:              device.KID.Bytes(),
			PublicKey:       device.PublicKey,
			AttestationType: device.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:       aaguid,
				SignCount:    device.SignCount,
				CloneWarning: device.CloneWarning,
			},
		}

		transports := strings.Split(device.Transport, ",")
		credential.Transport = make([]protocol.AuthenticatorTransport, len(transports))

		for _, t := range transports {
			transport := protocol.AuthenticatorTransport(t)

			switch transport {
			case protocol.Internal, protocol.USB, protocol.NFC, protocol.BLE:
				credential.Transport = append(credential.Transport, transport)
			}
		}

		credentials[i] = credential
	}

	return credentials
}

// WebAuthnCredentialDescriptors decodes the users credentials into protocol.CredentialDescriptor's.
func (w WebauthnUser) WebAuthnCredentialDescriptors() (descriptors []protocol.CredentialDescriptor) {
	descriptors = make([]protocol.CredentialDescriptor, len(w.Devices))

	for i, device := range w.Devices {
		descriptor := protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: device.KID.Bytes(),
		}

		for _, t := range strings.Split(device.Transport, ",") {
			transport := protocol.AuthenticatorTransport(t)

			switch transport {
			case protocol.Internal, protocol.USB, protocol.NFC, protocol.BLE:
				descriptor.Transport = append(descriptor.Transport, transport)
			}
		}

		descriptors[i] = descriptor
	}

	return descriptors
}

// NewWebauthnDeviceFromCredential creates a WebauthnDevice from a webauthn.Credential.
func NewWebauthnDeviceFromCredential(username, description string, ip net.IP, credential *webauthn.Credential) (device WebauthnDevice) {
	transport := make([]string, len(credential.Transport))

	for i, t := range credential.Transport {
		transport[i] = string(t)
	}

	device = WebauthnDevice{
		Username:        username,
		Created:         time.Now(),
		IP:              NewIP(ip),
		Description:     description,
		KID:             NewBase64(credential.ID),
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		SignCount:       credential.Authenticator.SignCount,
		CloneWarning:    credential.Authenticator.CloneWarning,
		Transport:       strings.Join(transport, ","),
	}

	device.AAGUID, _ = uuid.Parse(hex.EncodeToString(credential.Authenticator.AAGUID))

	return device
}

// WebauthnDevice represents a Webauthn Device in the database storage.
type WebauthnDevice struct {
	ID              int       `db:"id"`
	Created         time.Time `db:"created"`
	IP              IP        `db:"ip"`
	Username        string    `db:"username"`
	Description     string    `db:"description"`
	KID             Base64    `db:"kid"`
	PublicKey       []byte    `db:"public_key"`
	AttestationType string    `db:"attestation_type"`
	Transport       string    `db:"transport"`
	AAGUID          uuid.UUID `db:"aaguid"`
	SignCount       uint32    `db:"sign_count"`
	CloneWarning    bool      `db:"clone_warning"`
}
