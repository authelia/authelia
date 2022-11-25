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
type EncryptionCheckKeyFunc func(ctx context.Context, provider *SQLProvider) (err error)

type encOAuth2Session struct {
	ID      int    `db:"id"`
	Session []byte `db:"session_data"`
}

type encWebauthnDevice struct {
	ID        int    `db:"id"`
	PublicKey []byte `db:"public_key"`
}

type encTOTPConfiguration struct {
	ID     int    `db:"id" json:"-"`
	Secret []byte `db:"secret" json:"-"`
}
