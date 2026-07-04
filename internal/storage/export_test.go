package storage

import (
	"context"
	"crypto/sha256"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/utils"
)

// CtxKeyConnection is the exported context key for stashing a SQLXConnection used by tests in storage_test.
var CtxKeyConnection = ctxKeyConnection

// CtxKeyTransaction is the exported context key for stashing a SQLXTx used by tests in storage_test.
var CtxKeyTransaction = ctxKeyTransaction

// NewSQLProviderForTesting constructs an SQLProvider with the supplied db and a deterministic encryption key.
func NewSQLProviderForTesting(db SQLXDB) *SQLProvider {
	key, err := utils.DeriveCryptographicKey([]byte("test-encryption-key"), hkdfKeyInfo, sha256.New)
	if err != nil {
		panic(err)
	}

	return &SQLProvider{
		db:   db,
		name: providerSQLite,
		log:  logging.Logger(),
		keys: SQLProviderKeys{
			encryption: key,
		},
	}
}

// Conn exposes (*SQLProvider).conn for tests in storage_test.
func (p *SQLProvider) Conn(ctx context.Context) SQLXConnection {
	return p.conn(ctx)
}

// Encrypt exposes (*SQLProvider).encrypt for tests in storage_test.
func (p *SQLProvider) Encrypt(clearText, aad []byte) ([]byte, error) {
	return utils.Encrypt(clearText, aad, p.keys.encryption)
}
