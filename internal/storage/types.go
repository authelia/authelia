package storage

import (
	"context"

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

type encWebauthnDevice struct {
	ID        int    `db:"id"`
	PublicKey []byte `db:"public_key"`
}

type encTOTPConfiguration struct {
	ID     int    `db:"id"`
	Secret []byte `db:"secret"`
}

type encOneTimePassword struct {
	ID  int    `db:"id"`
	OTP []byte `db:"otp"`
}

type encEncryption struct {
	ID    int    `db:"id"`
	Value []byte `db:"value"`
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
