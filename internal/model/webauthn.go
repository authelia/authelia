package model

import (
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"go.yaml.in/yaml/v4"
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
func (u WebAuthnUser) HasFIDOU2F() bool {
	for _, c := range u.Credentials {
		if c.AttestationType == attestationTypeFIDOU2F {
			return true
		}
	}

	return false
}

// WebAuthnID implements the webauthn.User interface.
func (u WebAuthnUser) WebAuthnID() []byte {
	return []byte(u.UserID)
}

// WebAuthnName implements the webauthn.User  interface.
func (u WebAuthnUser) WebAuthnName() string {
	return u.Username
}

// WebAuthnDisplayName implements the webauthn.User interface.
func (u WebAuthnUser) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnIcon implements the webauthn.User interface.
func (u WebAuthnUser) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials implements the webauthn.User interface.
func (u WebAuthnUser) WebAuthnCredentials() (credentials []webauthn.Credential) {
	credentials = make([]webauthn.Credential, len(u.Credentials))

	var (
		credential *webauthn.Credential
		err        error
	)

	for i, c := range u.Credentials {
		if credential, err = c.ToCredential(); err != nil {
			continue
		}

		credentials[i] = *credential
	}

	return credentials
}

// WebAuthnCredentialDescriptors decodes the users credentials into protocol.CredentialDescriptor's.
func (u WebAuthnUser) WebAuthnCredentialDescriptors() (descriptors []protocol.CredentialDescriptor) {
	credentials := u.WebAuthnCredentials()

	descriptors = make([]protocol.CredentialDescriptor, len(credentials))

	for i, credential := range credentials {
		descriptors[i] = credential.Descriptor()
	}

	return descriptors
}

// NewWebAuthnCredential creates a WebAuthnCredential from a webauthn.Credential.
func NewWebAuthnCredential(ctx Context, rpid, username, description string, credential *webauthn.Credential) (c WebAuthnCredential) {
	transport := make([]string, len(credential.Transport))

	for i, t := range credential.Transport {
		transport[i] = string(t)
	}

	c = WebAuthnCredential{
		RPID:            rpid,
		Username:        username,
		CreatedAt:       ctx.GetClock().Now(),
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

	c.Attestation, _ = json.Marshal(credential.Attestation)

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
	Legacy          bool          `db:"legacy"`
	Discoverable    bool          `db:"discoverable"`
	Present         bool          `db:"present"`
	Verified        bool          `db:"verified"`
	BackupEligible  bool          `db:"backup_eligible"`
	BackupState     bool          `db:"backup_state"`
	PublicKey       []byte        `db:"public_key"`
	Attestation     []byte        `db:"attestation"`
}

// UpdateSignInInfo adjusts the values of the WebAuthnCredential after a sign in.
func (c *WebAuthnCredential) UpdateSignInInfo(config *webauthn.Config, now time.Time, authenticator webauthn.Authenticator) {
	c.LastUsedAt = sql.NullTime{Time: now, Valid: true}
	c.SignCount, c.CloneWarning = authenticator.SignCount, authenticator.CloneWarning

	if c.RPID != "" {
		return
	}

	switch c.AttestationType {
	case attestationTypeFIDOU2F:
		c.RPID = config.RPOrigins[0]
	default:
		c.RPID = config.RPID
	}
}

// DataValueLastUsedAt provides LastUsedAt as a *time.Time instead of sql.NullTime.
func (c *WebAuthnCredential) DataValueLastUsedAt() *time.Time {
	if c.LastUsedAt.Valid {
		value := time.Unix(c.LastUsedAt.Time.Unix(), int64(c.LastUsedAt.Time.Nanosecond()))

		return &value
	}

	return nil
}

// DataValueAAGUID provides AAGUID as a *string instead of uuid.NullUUID.
func (c *WebAuthnCredential) DataValueAAGUID() *string {
	if c.AAGUID.Valid {
		value := c.AAGUID.UUID.String()

		return &value
	}

	return nil
}

func (c *WebAuthnCredential) ToCredential() (credential *webauthn.Credential, err error) {
	credential = &webauthn.Credential{
		ID:              c.KID.Bytes(),
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Flags: webauthn.CredentialFlags{
			UserPresent:    c.Present,
			UserVerified:   c.Verified,
			BackupEligible: c.BackupEligible,
			BackupState:    c.BackupState,
		},
		Authenticator: webauthn.Authenticator{
			SignCount:    c.SignCount,
			CloneWarning: c.CloneWarning,
			Attachment:   protocol.AuthenticatorAttachment(c.Attachment),
		},
	}

	// This function never returns errors though we return here just in case that changes.
	if credential.Authenticator.AAGUID, err = c.AAGUID.MarshalBinary(); err != nil {
		return nil, err
	}

	if len(c.Attestation) != 0 {
		if err = json.Unmarshal(c.Attestation, &credential.Attestation); err != nil {
			return nil, err
		}
	}

	transports := strings.Split(c.Transport, ",")
	credential.Transport = []protocol.AuthenticatorTransport{}

	for _, t := range transports {
		if t == "" {
			continue
		}

		credential.Transport = append(credential.Transport, protocol.AuthenticatorTransport(t))
	}

	return credential, nil
}

func (c *WebAuthnCredential) ToData() WebAuthnCredentialData {
	o := WebAuthnCredentialData{
		ID:              c.ID,
		CreatedAt:       c.CreatedAt,
		LastUsedAt:      c.DataValueLastUsedAt(),
		RPID:            c.RPID,
		Username:        c.Username,
		Description:     c.Description,
		KID:             c.KID.String(),
		AAGUID:          c.DataValueAAGUID(),
		AttestationType: c.AttestationType,
		Attachment:      c.Attachment,
		SignCount:       c.SignCount,
		CloneWarning:    c.CloneWarning,
		Legacy:          c.Legacy,
		Discoverable:    c.Discoverable,
		Present:         c.Present,
		Verified:        c.Verified,
		BackupEligible:  c.BackupEligible,
		BackupState:     c.BackupState,
		PublicKey:       base64.StdEncoding.EncodeToString(c.PublicKey),
		Attestation:     base64.StdEncoding.EncodeToString(c.Attestation),
	}

	if c.Transport != "" {
		o.Transports = strings.Split(c.Transport, ",")
	}

	return o
}

// MarshalJSON returns the WebAuthnCredential in a JSON friendly manner.
func (c *WebAuthnCredential) MarshalJSON() (data []byte, err error) {
	return json.Marshal(c.ToData())
}

// MarshalYAML marshals this model into YAML.
func (c *WebAuthnCredential) MarshalYAML() (any, error) {
	return c.ToData(), nil
}

// UnmarshalYAML unmarshalls YAML into this model.
func (c *WebAuthnCredential) UnmarshalYAML(value *yaml.Node) (err error) {
	o := &WebAuthnCredentialData{}

	if err = value.Decode(o); err != nil {
		return err
	}

	if c.PublicKey, err = base64.StdEncoding.DecodeString(o.PublicKey); err != nil {
		return err
	}

	if len(o.Attestation) != 0 {
		if c.Attestation, err = base64.StdEncoding.DecodeString(o.Attestation); err != nil {
			return err
		}
	}

	var aaguid uuid.UUID

	if o.AAGUID != nil {
		if aaguid, err = uuid.Parse(*o.AAGUID); err != nil {
			return err
		}

		c.AAGUID = NullUUID(aaguid)
	}

	var kid []byte

	if kid, err = base64.StdEncoding.DecodeString(o.KID); err != nil {
		return err
	}

	c.KID = NewBase64(kid)

	c.CreatedAt = o.CreatedAt
	c.RPID = o.RPID
	c.Username = o.Username
	c.Description = o.Description
	c.AttestationType = o.AttestationType
	c.Attachment = o.Attachment
	c.Transport = strings.Join(o.Transports, ",")
	c.SignCount = o.SignCount
	c.CloneWarning = o.CloneWarning
	c.Discoverable = o.Discoverable
	c.Present = o.Present
	c.Verified = o.Verified
	c.BackupEligible = o.BackupEligible
	c.BackupState = o.BackupState

	if o.LastUsedAt != nil {
		c.LastUsedAt = sql.NullTime{Valid: true, Time: *o.LastUsedAt}
	}

	return nil
}

// WebAuthnCredentialData represents a WebAuthn Credential in a way which can be serialized.
type WebAuthnCredentialData struct {
	ID              int        `json:"id" yaml:"-"`
	CreatedAt       time.Time  `yaml:"created_at" json:"created_at" jsonschema:"title=Created At" jsonschema_description:"The time this credential was created."`
	LastUsedAt      *time.Time `yaml:"last_used_at,omitempty" json:"last_used_at,omitempty" jsonschema:"title=Last Used At" jsonschema_description:"The last time this credential was used."`
	RPID            string     `yaml:"rpid" json:"rpid" jsonschema:"title=Relying Party ID" jsonschema_description:"The Relying Party ID used to register this credential."`
	Username        string     `yaml:"username" json:"username" jsonschema:"title=Username" jsonschema_description:"The username of the user this credential belongs to."`
	Description     string     `yaml:"description" json:"description" jsonschema:"title=Description" jsonschema_description:"The user description of this credential."`
	KID             string     `yaml:"kid" json:"kid" jsonschema:"title=Public Key ID" jsonschema_description:"The Public Key ID of this credential."`
	AAGUID          *string    `yaml:"aaguid,omitempty" json:"aaguid,omitempty" jsonschema:"title=AAGUID" jsonschema_description:"The Authenticator Attestation Global Unique Identifier of this credential."`
	AttestationType string     `yaml:"attestation_type" json:"attestation_type" jsonschema:"title=Attestation Type" jsonschema_description:"The attestation format type this credential uses."`
	Attachment      string     `yaml:"attachment" json:"attachment" jsonschema:"title=Attachment" jsonschema_description:"The last recorded credential attachment type."`
	Transports      []string   `yaml:"transports" json:"transports" jsonschema:"title=Transports" jsonschema_description:"The last recorded credential transports."`
	SignCount       uint32     `yaml:"sign_count" json:"sign_count" jsonschema:"title=Sign Count" jsonschema_description:"The last recorded credential sign count."`
	CloneWarning    bool       `yaml:"clone_warning" json:"clone_warning" jsonschema:"title=Clone Warning" jsonschema_description:"The clone warning status of the credential."`
	Legacy          bool       `yaml:"legacy" json:"legacy" jsonschema:"title=Legacy" jsonschema_description:"The legacy value indicates this credential may need to be registered again."`
	Discoverable    bool       `yaml:"discoverable" json:"discoverable" jsonschema:"title=Discoverable" jsonschema_description:"The discoverable status of this credential."`
	Present         bool       `yaml:"present" json:"present" jsonschema:"title=Present" jsonschema_description:"The user presence status of this credential."`
	Verified        bool       `yaml:"verified" json:"verified" jsonschema:"title=Verified" jsonschema_description:"The verified status of this credential."`
	BackupEligible  bool       `yaml:"backup_eligible" json:"backup_eligible" jsonschema:"title=Backup Eligible" jsonschema_description:"The backup eligible status of this credential."`
	BackupState     bool       `yaml:"backup_state" json:"backup_state" jsonschema:"title=Backup Eligible" jsonschema_description:"The backup eligible status of this credential."`
	PublicKey       string     `yaml:"public_key" json:"public_key" jsonschema:"title=Public Key" jsonschema_description:"The credential public key."`
	Attestation     string     `yaml:"attestation" json:"attestation,omitempty" jsonschema:"title=Attestation" jsonschema_description:"The credential attestation information for auditing and validation."`
}

func (c *WebAuthnCredentialData) ToCredential() (credential *WebAuthnCredential, err error) {
	credential = &WebAuthnCredential{
		CreatedAt:       c.CreatedAt,
		RPID:            c.RPID,
		Username:        c.Username,
		Description:     c.Description,
		AttestationType: c.AttestationType,
		Attachment:      c.Attachment,
		Transport:       strings.Join(c.Transports, ","),
		SignCount:       c.SignCount,
		CloneWarning:    c.CloneWarning,
		Legacy:          c.Legacy,
		Discoverable:    c.Discoverable,
		Present:         c.Present,
		Verified:        c.Verified,
		BackupEligible:  c.BackupEligible,
		BackupState:     c.BackupState,
	}

	if len(c.PublicKey) != 0 {
		if credential.PublicKey, err = base64.StdEncoding.DecodeString(c.PublicKey); err != nil {
			return nil, err
		}
	}

	if len(c.Attestation) != 0 {
		if credential.Attestation, err = base64.StdEncoding.DecodeString(c.Attestation); err != nil {
			return nil, err
		}
	}

	var aaguid uuid.UUID

	if c.AAGUID != nil {
		if aaguid, err = uuid.Parse(*c.AAGUID); err != nil {
			return nil, fmt.Errorf("error occurred parsing aaguid: %w", err)
		}

		credential.AAGUID = NullUUID(aaguid)
	}

	if len(c.KID) != 0 {
		var kid []byte

		if kid, err = base64.StdEncoding.DecodeString(c.KID); err != nil {
			return nil, fmt.Errorf("error occurred deocding kid: %w", err)
		}

		credential.KID = NewBase64(kid)
	}

	if c.LastUsedAt != nil {
		credential.LastUsedAt = sql.NullTime{Valid: true, Time: *c.LastUsedAt}
	}

	return credential, nil
}

// WebAuthnCredentialExport represents a WebAuthnCredential export file.
type WebAuthnCredentialExport struct {
	WebAuthnCredentials []WebAuthnCredential `yaml:"webauthn_credentials"`
}

// WebAuthnCredentialDataExport represents a WebAuthnCredential export file.
type WebAuthnCredentialDataExport struct {
	WebAuthnCredentials []WebAuthnCredentialData `yaml:"webauthn_credentials" json:"webauthn_credentials" jsonschema:"title=WebAuthn Credentials" jsonschema_description:"The list of WebAuthn credentials."`
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
