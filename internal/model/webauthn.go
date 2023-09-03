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
	yaml "gopkg.in/yaml.v3"
)

const (
	attestationTypeFIDOU2F = "fido-u2f"
)

// WebAuthnUser is an object to represent a user for the WebAuthn lib.
type WebAuthnUser struct {
	ID          int    `db:"id"`
	RPID        string `db:"rpid"`
	Username    string `db:"username"`
	UserID      string `db:"userid"`
	DisplayName string `db:"-"`

	Credentials []WebAuthnCredential `db:"-"`
}

// HasFIDOU2F returns true if the user has any attestation type `fido-u2f` credentials.
func (w WebAuthnUser) HasFIDOU2F() bool {
	for _, c := range w.Credentials {
		if c.AttestationType == attestationTypeFIDOU2F {
			return true
		}
	}

	return false
}

// WebAuthnID implements the webauthn.User interface.
func (w WebAuthnUser) WebAuthnID() []byte {
	return []byte(w.UserID)
}

// WebAuthnName implements the webauthn.User  interface.
func (w WebAuthnUser) WebAuthnName() string {
	return w.Username
}

// WebAuthnDisplayName implements the webauthn.User interface.
func (w WebAuthnUser) WebAuthnDisplayName() string {
	return w.DisplayName
}

// WebAuthnIcon implements the webauthn.User interface.
func (w WebAuthnUser) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials implements the webauthn.User interface.
func (w WebAuthnUser) WebAuthnCredentials() (credentials []webauthn.Credential) {
	credentials = make([]webauthn.Credential, len(w.Credentials))

	var c webauthn.Credential

	for i, credential := range w.Credentials {
		aaguid, err := credential.AAGUID.MarshalBinary()
		if err != nil {
			continue
		}

		c = webauthn.Credential{
			ID:              credential.KID.Bytes(),
			PublicKey:       credential.PublicKey,
			AttestationType: credential.AttestationType,
			Flags: webauthn.CredentialFlags{
				UserPresent:    credential.Present,
				UserVerified:   credential.Verified,
				BackupEligible: credential.BackupEligible,
				BackupState:    credential.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				AAGUID:       aaguid,
				SignCount:    credential.SignCount,
				CloneWarning: credential.CloneWarning,
				Attachment:   protocol.AuthenticatorAttachment(credential.Attachment),
			},
		}

		transports := strings.Split(credential.Transport, ",")
		c.Transport = []protocol.AuthenticatorTransport{}

		for _, t := range transports {
			if t == "" {
				continue
			}

			c.Transport = append(c.Transport, protocol.AuthenticatorTransport(t))
		}

		credentials[i] = c
	}

	return credentials
}

// WebAuthnCredentialDescriptors decodes the users credentials into protocol.CredentialDescriptor's.
func (w WebAuthnUser) WebAuthnCredentialDescriptors() (descriptors []protocol.CredentialDescriptor) {
	credentials := w.WebAuthnCredentials()

	descriptors = make([]protocol.CredentialDescriptor, len(credentials))

	for i, credential := range credentials {
		descriptors[i] = credential.Descriptor()
	}

	return descriptors
}

// NewWebAuthnCredential creates a WebAuthnCredential from a webauthn.Credential.
func NewWebAuthnCredential(rpid, username, description string, credential *webauthn.Credential) (c WebAuthnCredential) {
	transport := make([]string, len(credential.Transport))

	for i, t := range credential.Transport {
		transport[i] = string(t)
	}

	c = WebAuthnCredential{
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
	if err == nil {
		c.AAGUID = NullUUID(aaguid)
	}

	return c
}

// WebAuthnCredential represents a WebAuthn Credential in the database storage.
type WebAuthnCredential struct {
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

// UpdateSignInInfo adjusts the values of the WebAuthnCredential after a sign in.
func (d *WebAuthnCredential) UpdateSignInInfo(config *webauthn.Config, now time.Time, signCount uint32) {
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

// DataValueLastUsedAt provides LastUsedAt as a *time.Time instead of sql.NullTime.
func (d *WebAuthnCredential) DataValueLastUsedAt() *time.Time {
	if d.LastUsedAt.Valid {
		value := time.Unix(d.LastUsedAt.Time.Unix(), int64(d.LastUsedAt.Time.Nanosecond()))

		return &value
	}

	return nil
}

// DataValueAAGUID provides AAGUID as a *string instead of uuid.NullUUID.
func (d *WebAuthnCredential) DataValueAAGUID() *string {
	if d.AAGUID.Valid {
		value := d.AAGUID.UUID.String()

		return &value
	}

	return nil
}

func (d *WebAuthnCredential) ToData() WebAuthnCredentialData {
	o := WebAuthnCredentialData{
		ID:              d.ID,
		CreatedAt:       d.CreatedAt,
		LastUsedAt:      d.DataValueLastUsedAt(),
		RPID:            d.RPID,
		Username:        d.Username,
		Description:     d.Description,
		KID:             d.KID.String(),
		AAGUID:          d.DataValueAAGUID(),
		AttestationType: d.AttestationType,
		Attachment:      d.Attachment,
		SignCount:       d.SignCount,
		CloneWarning:    d.CloneWarning,
		Present:         d.Present,
		Verified:        d.Verified,
		BackupEligible:  d.BackupEligible,
		BackupState:     d.BackupState,
		PublicKey:       base64.StdEncoding.EncodeToString(d.PublicKey),
	}

	if d.Transport != "" {
		o.Transports = strings.Split(d.Transport, ",")
	}

	return o
}

// MarshalJSON returns the WebAuthnCredential in a JSON friendly manner.
func (d *WebAuthnCredential) MarshalJSON() (data []byte, err error) {
	return json.Marshal(d.ToData())
}

// MarshalYAML marshals this model into YAML.
func (d *WebAuthnCredential) MarshalYAML() (any, error) {
	return d.ToData(), nil
}

// UnmarshalYAML unmarshalls YAML into this model.
func (d *WebAuthnCredential) UnmarshalYAML(value *yaml.Node) (err error) {
	o := &WebAuthnCredentialData{}

	if err = value.Decode(o); err != nil {
		return err
	}

	if d.PublicKey, err = base64.StdEncoding.DecodeString(o.PublicKey); err != nil {
		return err
	}

	var aaguid uuid.UUID

	if o.AAGUID != nil {
		if aaguid, err = uuid.Parse(*o.AAGUID); err != nil {
			return err
		}

		d.AAGUID = NullUUID(aaguid)
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
	d.Transport = strings.Join(o.Transports, ",")
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

// WebAuthnCredentialData represents a WebAuthn Credential in a way which can be serialized.
type WebAuthnCredentialData struct {
	ID              int        `json:"id" yaml:"-"`
	CreatedAt       time.Time  `yaml:"created_at" json:"created_at" jsonschema:"title=Created At" jsonschema_description:"The time this credential was created"`
	LastUsedAt      *time.Time `yaml:"last_used_at,omitempty" json:"last_used_at,omitempty" jsonschema:"title=Last Used At" jsonschema_description:"The last time this credential was used"`
	RPID            string     `yaml:"rpid" json:"rpid" jsonschema:"title=Relying Party ID" jsonschema_description:"The Relying Party ID used to register this credential"`
	Username        string     `yaml:"username" json:"username" jsonschema:"title=Username" jsonschema_description:"The username of the user this credential belongs to"`
	Description     string     `yaml:"description" json:"description" jsonschema:"title=Description" jsonschema_description:"The user description of this credential"`
	KID             string     `yaml:"kid" json:"kid" jsonschema:"title=Public Key ID" jsonschema_description:"The Public Key ID of this credential"`
	AAGUID          *string    `yaml:"aaguid,omitempty" json:"aaguid,omitempty" jsonschema:"title=AAGUID" jsonschema_description:"The Authenticator Attestation Global Unique Identifier of this credential"`
	AttestationType string     `yaml:"attestation_type" json:"attestation_type" jsonschema:"title=Attestation Type" jsonschema_description:"The attestation format type this credential uses"`
	Attachment      string     `yaml:"attachment" json:"attachment" jsonschema:"title=Attachment" jsonschema_description:"The last recorded credential attachment type"`
	Transports      []string   `yaml:"transports" json:"transports" jsonschema:"title=Transports" jsonschema_description:"The last recorded credential transports"`
	SignCount       uint32     `yaml:"sign_count" json:"sign_count" jsonschema:"title=Sign Count" jsonschema_description:"The last recorded credential sign count"`
	CloneWarning    bool       `yaml:"clone_warning" json:"clone_warning" jsonschema:"title=Clone Warning" jsonschema_description:"The clone warning status of the credential"`
	Discoverable    bool       `yaml:"discoverable" json:"discoverable" jsonschema:"title=Discoverable" jsonschema_description:"The discoverable status of this credential"`
	Present         bool       `yaml:"present" json:"present" jsonschema:"title=Present" jsonschema_description:"The user presence status of this credential"`
	Verified        bool       `yaml:"verified" json:"verified" jsonschema:"title=Verified" jsonschema_description:"The verified status of this credential"`
	BackupEligible  bool       `yaml:"backup_eligible" json:"backup_eligible" jsonschema:"title=Backup Eligible" jsonschema_description:"The backup eligible status of this credential"`
	BackupState     bool       `yaml:"backup_state" json:"backup_state" jsonschema:"title=Backup Eligible" jsonschema_description:"The backup eligible status of this credential"`
	PublicKey       string     `yaml:"public_key" json:"public_key" jsonschema:"title=Public Key" jsonschema_description:"The credential public key"`
}

func (d *WebAuthnCredentialData) ToCredential() (credential *WebAuthnCredential, err error) {
	credential = &WebAuthnCredential{
		CreatedAt:       d.CreatedAt,
		RPID:            d.RPID,
		Username:        d.Username,
		Description:     d.Description,
		AttestationType: d.AttestationType,
		Attachment:      d.Attachment,
		Transport:       strings.Join(d.Transports, ","),
		SignCount:       d.SignCount,
		CloneWarning:    d.CloneWarning,
		Discoverable:    d.Discoverable,
		Present:         d.Present,
		Verified:        d.Verified,
		BackupEligible:  d.BackupEligible,
		BackupState:     d.BackupState,
	}

	if credential.PublicKey, err = base64.StdEncoding.DecodeString(d.PublicKey); err != nil {
		return nil, err
	}

	var aaguid uuid.UUID

	if d.AAGUID != nil {
		if aaguid, err = uuid.Parse(*d.AAGUID); err != nil {
			return nil, err
		}

		credential.AAGUID = NullUUID(aaguid)
	}

	var kid []byte

	if kid, err = base64.StdEncoding.DecodeString(d.KID); err != nil {
		return nil, err
	}

	credential.KID = NewBase64(kid)

	if d.LastUsedAt != nil {
		credential.LastUsedAt = sql.NullTime{Valid: true, Time: *d.LastUsedAt}
	}

	return credential, nil
}

// WebAuthnCredentialExport represents a WebAuthnCredential export file.
type WebAuthnCredentialExport struct {
	WebAuthnCredentials []WebAuthnCredential `yaml:"webauthn_credentials"`
}

// WebAuthnCredentialDataExport represents a WebAuthnCredential export file.
type WebAuthnCredentialDataExport struct {
	WebAuthnCredentials []WebAuthnCredentialData `yaml:"webauthn_credentials" json:"webauthn_credentials" jsonschema:"title=WebAuthn Credentials" jsonschema_description:"The list of WebAuthn credentials"`
}

// ToData converts this WebAuthnCredentialExport into a WebAuthnCredentialDataExport.
func (export WebAuthnCredentialExport) ToData() WebAuthnCredentialDataExport {
	data := WebAuthnCredentialDataExport{
		WebAuthnCredentials: make([]WebAuthnCredentialData, len(export.WebAuthnCredentials)),
	}

	for i, credential := range export.WebAuthnCredentials {
		data.WebAuthnCredentials[i] = credential.ToData()
	}

	return data
}

// MarshalYAML marshals this model into YAML.
func (export WebAuthnCredentialExport) MarshalYAML() (any, error) {
	return export.ToData(), nil
}
