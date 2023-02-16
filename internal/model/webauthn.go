package model

import (
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

const (
	attestationTypeFIDOU2F = "fido-u2f"
)

// WebauthnUser is an object to represent a user for the Webauthn lib.
type WebauthnUser struct {
	ID          int    `db:"id"`
	RPID        string `db:"rpid"`
	Username    string `db:"username"`
	UserID      string `db:"userid"`
	DisplayName string `db:"-"`

	Devices []WebauthnDevice `db:"-"`
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
	return []byte(w.UserID)
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
			Flags: webauthn.CredentialFlags{
				UserPresent:    device.Present,
				UserVerified:   device.Verified,
				BackupEligible: device.BackupEligible,
				BackupState:    device.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				AAGUID:       aaguid,
				SignCount:    device.SignCount,
				CloneWarning: device.CloneWarning,
				Attachment:   protocol.AuthenticatorAttachment(device.Attachment),
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
		AttestationType: credential.AttestationType,
		Attachment:      string(credential.Authenticator.Attachment),
		Transport:       strings.Join(transport, ","),
		SignCount:       credential.Authenticator.SignCount,
		CloneWarning:    credential.Authenticator.CloneWarning,
		Discoverable:    false,
		Present:         credential.Flags.UserPresent,
		Verified:        credential.Flags.UserVerified,
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
		PublicKey:       credential.PublicKey,
	}

	aaguid, err := uuid.Parse(hex.EncodeToString(credential.Authenticator.AAGUID))
	if err == nil && aaguid.ID() != 0 {
		device.AAGUID = uuid.NullUUID{Valid: true, UUID: aaguid}
	}

	return device
}

// WebauthnDeviceJSON represents a Webauthn Device in the JSON format.
type WebauthnDeviceJSON struct {
	ID              int        `json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	RPID            string     `json:"rpid"`
	Description     string     `json:"description"`
	KID             []byte     `json:"kid"`
	AAGUID          string     `json:"aaguid,omitempty"`
	Attachment      string     `json:"attachment"`
	AttestationType string     `json:"attestation_type"`
	Transports      []string   `json:"transports"`
	SignCount       uint32     `json:"sign_count"`
	CloneWarning    bool       `json:"clone_warning"`
	Discoverable    bool       `json:"discoverable"`
	Present         bool       `json:"present"`
	Verified        bool       `json:"verified"`
	BackupEligible  bool       `json:"backup_eligible"`
	BackupState     bool       `json:"backup_state"`
	PublicKey       []byte     `json:"public_key"`
}

// WebauthnDevice represents a Webauthn Device in the database storage.
type WebauthnDevice struct {
	ID              int           `db:"id"`
	CreatedAt       time.Time     `db:"created_at"`
	LastUsedAt      sql.NullTime  `db:"last_used_at"`
	RPID            string        `db:"rpid"`
	Username        string        `db:"username"`
	Description     string        `db:"description"`
	KID             Base64        `db:"kid"`
	AAGUID          uuid.NullUUID `db:"aaguid"`
	AttestationType string        `db:"attestation_type"`
	Attachment      string        `db:"attachment"`
	Transport       string        `db:"transport"`
	SignCount       uint32        `db:"sign_count"`
	CloneWarning    bool          `db:"clone_warning"`
	Discoverable    bool          `db:"discoverable"`
	Present         bool          `db:"present"`
	Verified        bool          `db:"verified"`
	BackupEligible  bool          `db:"backup_eligible"`
	BackupState     bool          `db:"backup_state"`
	PublicKey       []byte        `db:"public_key"`
}

// MarshalJSON returns the WebauthnDevice in a JSON friendly manner.
func (w *WebauthnDevice) MarshalJSON() (data []byte, err error) {
	o := WebauthnDeviceJSON{
		ID:              w.ID,
		CreatedAt:       w.CreatedAt,
		RPID:            w.RPID,
		Description:     w.Description,
		KID:             w.KID.data,
		AttestationType: w.AttestationType,
		Attachment:      w.Attachment,
		Transports:      []string{},
		SignCount:       w.SignCount,
		CloneWarning:    w.CloneWarning,
		Discoverable:    w.Discoverable,
		Present:         w.Present,
		Verified:        w.Verified,
		BackupEligible:  w.BackupEligible,
		BackupState:     w.BackupState,
		PublicKey:       w.PublicKey,
	}

	if w.AAGUID.Valid {
		o.AAGUID = w.AAGUID.UUID.String()
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
func (d *WebauthnDevice) UpdateSignInInfo(config *webauthn.Config, now time.Time, signCount uint32) {
	d.LastUsedAt = sql.NullTime{Time: now, Valid: true}

	d.SignCount = signCount

	if d.RPID != "" {
		return
	}

	switch d.AttestationType {
	case attestationTypeFIDOU2F:
		d.RPID = config.RPOrigins[0]
	default:
		d.RPID = config.RPID
	}
}

func (d *WebauthnDevice) LastUsed() *time.Time {
	if d.LastUsedAt.Valid {
		return &d.LastUsedAt.Time
	}

	return nil
}

// MarshalYAML marshals this model into YAML.
func (d *WebauthnDevice) MarshalYAML() (any, error) {
	o := WebauthnDeviceData{
		CreatedAt:       d.CreatedAt,
		LastUsedAt:      d.LastUsed(),
		RPID:            d.RPID,
		Username:        d.Username,
		Description:     d.Description,
		KID:             d.KID.String(),
		AAGUID:          d.AAGUID.UUID.String(),
		AttestationType: d.AttestationType,
		Attachment:      d.Attachment,
		Transport:       d.Transport,
		SignCount:       d.SignCount,
		CloneWarning:    d.CloneWarning,
		Present:         d.Present,
		Verified:        d.Verified,
		BackupEligible:  d.BackupEligible,
		BackupState:     d.BackupState,
		PublicKey:       base64.StdEncoding.EncodeToString(d.PublicKey),
	}

	return yaml.Marshal(o)
}

// UnmarshalYAML unmarshalls YAML into this model.
func (d *WebauthnDevice) UnmarshalYAML(value *yaml.Node) (err error) {
	o := &WebauthnDeviceData{}

	if err = value.Decode(o); err != nil {
		return err
	}

	if d.PublicKey, err = base64.StdEncoding.DecodeString(o.PublicKey); err != nil {
		return err
	}

	var aaguid uuid.UUID

	if aaguid, err = uuid.Parse(o.AAGUID); err != nil {
		return err
	}

	if aaguid.ID() != 0 {
		d.AAGUID = uuid.NullUUID{Valid: true, UUID: aaguid}
	}

	var kid []byte

	if kid, err = base64.StdEncoding.DecodeString(o.KID); err != nil {
		return err
	}

	d.KID = NewBase64(kid)

	d.CreatedAt = o.CreatedAt
	d.RPID = o.RPID
	d.Username = o.Username
	d.Description = o.Description
	d.AttestationType = o.AttestationType
	d.Attachment = o.Attachment
	d.Transport = o.Transport
	d.SignCount = o.SignCount
	d.CloneWarning = o.CloneWarning
	d.Discoverable = o.Discoverable
	d.Present = o.Present
	d.Verified = o.Verified
	d.BackupEligible = o.BackupEligible
	d.BackupState = o.BackupState

	if o.LastUsedAt != nil {
		d.LastUsedAt = sql.NullTime{Valid: true, Time: *o.LastUsedAt}
	}

	return nil
}

// WebauthnDeviceData represents a Webauthn Device in the database storage.
type WebauthnDeviceData struct {
	CreatedAt       time.Time  `yaml:"created_at"`
	LastUsedAt      *time.Time `yaml:"last_used_at"`
	RPID            string     `yaml:"rpid"`
	Username        string     `yaml:"username"`
	Description     string     `yaml:"description"`
	KID             string     `yaml:"kid"`
	AAGUID          string     `yaml:"aaguid"`
	AttestationType string     `yaml:"attestation_type"`
	Attachment      string     `yaml:"attachment"`
	Transport       string     `yaml:"transport"`
	SignCount       uint32     `yaml:"sign_count"`
	CloneWarning    bool       `yaml:"clone_warning"`
	Discoverable    bool       `yaml:"discoverable"`
	Present         bool       `yaml:"present"`
	Verified        bool       `yaml:"verified"`
	BackupEligible  bool       `yaml:"backup_eligible"`
	BackupState     bool       `yaml:"backup_state"`
	PublicKey       string     `yaml:"public_key"`
}

// WebauthnDeviceExport represents a WebauthnDevice export file.
type WebauthnDeviceExport struct {
	WebauthnDevices []WebauthnDevice `yaml:"webauthn_devices"`
}
