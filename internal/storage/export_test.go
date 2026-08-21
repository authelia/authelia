package storage

import (
	"context"
	"crypto/sha256"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ProviderMySQL is the exported MySQL provider name used by tests in storage_test.
const ProviderMySQL = providerMySQL

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
		aad: aadRow,
	}
}

// NewSQLProviderForTestingWithKey constructs an SQLProvider with the supplied db and encryption key. It's used by
// tests that need to exercise the encryption failure paths by supplying an invalid key.
func NewSQLProviderForTestingWithKey(db SQLXDB, key []byte) *SQLProvider {
	return &SQLProvider{
		db:   db,
		name: providerSQLite,
		log:  logging.Logger(),
		keys: SQLProviderKeys{
			encryption: key,
		},
		aad: aadRow,
	}
}

// NewSQLProviderForTestingWithName constructs an SQLProvider with the supplied db and provider name. It's used by
// tests that need to exercise the provider specific branches such as the MySQL only special migration transaction.
func NewSQLProviderForTestingWithName(db SQLXDB, name string) *SQLProvider {
	p := NewSQLProviderForTesting(db)

	p.name = name
	p.sqlInsertMigration = "INSERT INTO migrations"

	return p
}

// Conn exposes (*SQLProvider).conn for tests in storage_test.
func (p *SQLProvider) Conn(ctx context.Context) SQLXConnection {
	return p.conn(ctx)
}

// SchemaMigrateApply exposes (*SQLProvider).schemaMigrateApply for tests in storage_test.
func (p *SQLProvider) SchemaMigrateApply(ctx context.Context, conn SQLXConnection, migration model.SchemaMigration, prior, target int) error {
	return p.schemaMigrateApply(ctx, conn, migration, prior, target)
}

// Encrypt exposes (*SQLProvider).encrypt for tests in storage_test.
func (p *SQLProvider) Encrypt(clearText, aad []byte) ([]byte, error) {
	return utils.Encrypt(clearText, aad, p.keys.encryption)
}
