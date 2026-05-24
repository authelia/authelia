package storage

import (
	"context"
	"crypto/sha256"

	"github.com/authelia/authelia/v4/internal/logging"
)

// CtxKeyConnection is the exported context key for stashing a SQLXConnection used by tests in storage_test.
var CtxKeyConnection = ctxKeyConnection

// CtxKeyTransaction is the exported context key for stashing a SQLXTx used by tests in storage_test.
var CtxKeyTransaction = ctxKeyTransaction

// ProviderSQLite is the exported provider name used by tests in storage_test.
const ProviderSQLite = providerSQLite

// NewSQLProviderForTesting constructs an SQLProvider with the supplied db and a deterministic encryption key.
func NewSQLProviderForTesting(db SQLXDB) *SQLProvider {
	return &SQLProvider{
		db:   db,
		name: providerSQLite,
		log:  logging.Logger(),
		keys: SQLProviderKeys{
			encryption: sha256.Sum256([]byte("test-encryption-key")),
		},
	}
}

// Conn exposes (*SQLProvider).conn for tests in storage_test.
func (p *SQLProvider) Conn(ctx context.Context) SQLXConnection {
	return p.conn(ctx)
}

// WithOpenErr sets the errOpen field on the SQLProvider for tests in storage_test.
func (p *SQLProvider) WithOpenErr(err error) *SQLProvider {
	p.errOpen = err
	return p
}

// Encrypt exposes (*SQLProvider).encrypt for tests in storage_test.
func (p *SQLProvider) Encrypt(clearText []byte) ([]byte, error) {
	return p.encrypt(clearText)
}
