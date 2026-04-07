package storage

import (
	"context"

	"github.com/authelia/authelia/v4/internal/utils"
)

type fSchemaMigration func(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, target int) (err error)

var migrationsSpecialUp = map[int][]fSchemaMigration{
	24: {migrationSpecialUp24},
}

var migrationsSpecialDown = map[int][]fSchemaMigration{
	24: {migrationSpecialDown24},
}

func migrationSpecialUp24(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, target int) (err error) {
	decryptKey := utils.DeriveLegacyEncryptionKey([]byte(provider.config.Storage.EncryptionKey))

	return schemaEncryptionChangeKey(ctx, conn, decryptKey, provider.keys.encryption)
}

func migrationSpecialDown24(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, target int) (err error) {
	encryptKey := utils.DeriveLegacyEncryptionKey([]byte(provider.config.Storage.EncryptionKey))

	return schemaEncryptionChangeKey(ctx, conn, provider.keys.encryption, encryptKey)
}
