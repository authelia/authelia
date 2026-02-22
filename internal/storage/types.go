package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// SQLXConnection is a *sqlx.DB or *sqlx.Tx.
type SQLXConnection interface {
	sqlx.Execer
	sqlx.ExecerContext

	sqlx.Preparer
	sqlx.PreparerContext

	sqlx.Queryer
	sqlx.QueryerContext

	sqlx.Ext
	sqlx.ExtContext
}

// EncryptionChangeKeyFunc handles encryption key changes for a specific table or tables.
type EncryptionChangeKeyFunc func(ctx context.Context, provider *SQLProvider, tx *sqlx.Tx, key [32]byte) (err error)

// EncryptionCheckKeyFunc handles encryption key checking for a specific table or tables.
type EncryptionCheckKeyFunc func(ctx context.Context, provider *SQLProvider) (table string, result EncryptionValidationTableResult)

type encOAuth2Session struct {
	ID      int    `db:"id"`
	Session []byte `db:"session_data"`
}

type encWebAuthnCredential struct {
	ID          int    `db:"id"`
	PublicKey   []byte `db:"public_key"`
	Attestation []byte `db:"attestation"`
}

type encCachedData struct {
	ID    int    `db:"id"`
	Value []byte `db:"value"`
}

type encTOTPConfiguration struct {
	ID     int    `db:"id"`
	Secret []byte `db:"secret"` //nolint:gosec
}

type encOneTimeCode struct {
	ID   int    `db:"id"`
	Code []byte `db:"code"`
}

type encEncryption struct {
	ID    int    `db:"id"`
	Value []byte `db:"value"`
}

type banExpiresExpired struct {
	Expires sql.NullTime `db:"expires"`
	Expired sql.NullTime `db:"expired"`
	Revoked bool         `db:"revoked"`
}

func (b *banExpiresExpired) Expiration() time.Time {
	if b.Revoked && b.Expired.Valid {
		return b.Expired.Time
	}

	if b.Expires.Valid {
		return b.Expires.Time
	}

	return time.Unix(0, 0)
}

// EncryptionValidationResult contains information about the success of a schema encryption validation.
type EncryptionValidationResult struct {
	InvalidCheckValue bool
	Tables            map[string]EncryptionValidationTableResult
}

// Success returns true if no validation errors occurred.
func (r EncryptionValidationResult) Success() bool {
	if r.InvalidCheckValue {
		return false
	}

	for _, table := range r.Tables {
		if table.Invalid != 0 || table.Error != nil {
			return false
		}
	}

	return true
}

// Checked returns true the validation completed all phases even if there were errors.
func (r EncryptionValidationResult) Checked() bool {
	for _, table := range r.Tables {
		if table.Error != nil {
			return false
		}
	}

	return true
}

// EncryptionValidationTableResult contains information about the success of a table schema encryption validation.
type EncryptionValidationTableResult struct {
	Error   error
	Total   int
	Invalid int
}

// ResultDescriptor returns a string representing the result.
func (r EncryptionValidationTableResult) ResultDescriptor() string {
	if r.Total == 0 {
		return na
	}

	if r.Error != nil || r.Invalid != 0 {
		return "FAILURE"
	}

	return "SUCCESS"
}

// OAuth2SessionType represents the potential OAuth 2.0 session types.
type OAuth2SessionType int

// Representation of specific OAuth 2.0 session types.
const (
	OAuth2SessionTypeAccessToken OAuth2SessionType = iota
	OAuth2SessionTypeAuthorizeCode
	OAuth2SessionTypeDeviceAuthorizeCode
	OAuth2SessionTypeOpenIDConnect
	OAuth2SessionTypePAR
	OAuth2SessionTypePKCEChallenge
	OAuth2SessionTypeRefreshToken
)

// String returns a string representation of this OAuth2SessionType.
func (s OAuth2SessionType) String() string {
	switch s {
	case OAuth2SessionTypeAccessToken:
		return "access token"
	case OAuth2SessionTypeAuthorizeCode:
		return "authorization code"
	case OAuth2SessionTypeDeviceAuthorizeCode:
		return "device code"
	case OAuth2SessionTypeOpenIDConnect:
		return "openid connect"
	case OAuth2SessionTypePAR:
		return "pushed authorization request context"
	case OAuth2SessionTypePKCEChallenge:
		return "pkce challenge"
	case OAuth2SessionTypeRefreshToken:
		return "refresh token"
	default:
		return "invalid"
	}
}

// Table returns the table name for this session type.
func (s OAuth2SessionType) Table() string {
	switch s {
	case OAuth2SessionTypeAccessToken:
		return tableOAuth2AccessTokenSession
	case OAuth2SessionTypeAuthorizeCode:
		return tableOAuth2AuthorizeCodeSession
	case OAuth2SessionTypeDeviceAuthorizeCode:
		return tableOAuth2DeviceCodeSession
	case OAuth2SessionTypeOpenIDConnect:
		return tableOAuth2OpenIDConnectSession
	case OAuth2SessionTypePAR:
		return tableOAuth2PARContext
	case OAuth2SessionTypePKCEChallenge:
		return tableOAuth2PKCERequestSession
	case OAuth2SessionTypeRefreshToken:
		return tableOAuth2RefreshTokenSession
	default:
		return ""
	}
}
