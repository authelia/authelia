package model

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

const (
	attestationTypeFIDOU2F = "fido-u2f"
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
		if c.AttestationType == attestationTypeFIDOU2F {
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
		credential.Transport = []protocol.AuthenticatorTransport{}

		for _, t := range transports {
			if t == "" {
				continue
			}

			credential.Transport = append(credential.Transport, protocol.AuthenticatorTransport(t))
		}

		credentials[i] = credential
	}

	return credentials
}

// WebAuthnCredentialDescriptors decodes the users credentials into protocol.CredentialDescriptor's.
func (w WebauthnUser) WebAuthnCredentialDescriptors() (descriptors []protocol.CredentialDescriptor) {
	credentials := w.WebAuthnCredentials()

	descriptors = make([]protocol.CredentialDescriptor, len(credentials))

	for i, credential := range credentials {
		descriptors[i] = credential.Descriptor()
	}

	return descriptors
}

// NewWebauthnDeviceFromCredential creates a WebauthnDevice from a webauthn.Credential.
func NewWebauthnDeviceFromCredential(rpid, username, description string, credential *webauthn.Credential) (device WebauthnDevice) {
	transport := make([]string, len(credential.Transport))

	for i, t := range credential.Transport {
		transport[i] = string(t)
	}

	device = WebauthnDevice{
		RPID:            rpid,
		Username:        username,
		CreatedAt:       time.Now(),
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
	ID              int          `db:"id"`
	CreatedAt       time.Time    `db:"created_at"`
	LastUsedAt      sql.NullTime `db:"last_used_at"`
	RPID            string       `db:"rpid"`
	Username        string       `db:"username"`
	Description     string       `db:"description"`
	KID             Base64       `db:"kid"`
	PublicKey       []byte       `db:"public_key"`
	AttestationType string       `db:"attestation_type"`
	Transport       string       `db:"transport"`
	AAGUID          uuid.UUID    `db:"aaguid"`
	SignCount       uint32       `db:"sign_count"`
	CloneWarning    bool         `db:"clone_warning"`
}

// MarshalJSON returns the WebauthnDevice in a JSON friendly manner.
func (w *WebauthnDevice) MarshalJSON() (data []byte, err error) {
	o := struct {
		ID              int        `json:"id"`
		CreatedAt       time.Time  `json:"created_at"`
		LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
		RPID            string     `json:"rpid"`
		Description     string     `json:"description"`
		KID             []byte     `json:"kid"`
		PublicKey       []byte     `json:"public_key"`
		AttestationType string     `json:"attestation_type"`
		Transports      []string   `json:"transports"`
		AAGUID          string     `json:"aaguid"`
		SignCount       uint32     `json:"sign_count"`
		CloneWarning    bool       `json:"clone_warning"`
	}{
		ID:              w.ID,
		CreatedAt:       w.CreatedAt,
		RPID:            w.RPID,
		Description:     w.Description,
		KID:             w.KID.data,
		PublicKey:       w.PublicKey,
		AttestationType: w.AttestationType,
		Transports:      []string{},
		AAGUID:          w.AAGUID.String(),
		SignCount:       w.SignCount,
		CloneWarning:    w.CloneWarning,
	}

	if w.Transport != "" {
		o.Transports = strings.Split(w.Transport, ",")
	}

	if w.LastUsedAt.Valid {
		o.LastUsedAt = &w.LastUsedAt.Time
	}

	return json.Marshal(o)
}

// UpdateSignInInfo adjusts the values of the WebauthnDevice after a sign in.
func (w *WebauthnDevice) UpdateSignInInfo(config *webauthn.Config, now time.Time, signCount uint32) {
	w.LastUsedAt = sql.NullTime{Time: now, Valid: true}

	w.SignCount = signCount

	if w.RPID != "" {
		return
	}

	switch w.AttestationType {
	case attestationTypeFIDOU2F:
		w.RPID = config.RPOrigin
	default:
		w.RPID = config.RPID
	}
}
