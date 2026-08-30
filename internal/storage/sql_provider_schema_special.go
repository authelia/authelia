package storage

import (
	"context"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

type fSchemaMigration func(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, before, after, target int) (err error)

var migrationsSpecialUp = map[int][]fSchemaMigration{
	24: {migrationSpecialUp24},
	25: {migrationSpecialUp25},
	26: {migrationSpecialUp26},
}

var migrationsSpecialDown = map[int][]fSchemaMigration{
	25: {migrationSpecialDown25},
	26: {migrationSpecialDown26},
}

func migrationSpecialUp24(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, before, after, target int) (err error) {
	var (
		credentials []model.WebAuthnCredential
		credential  *webauthn.Credential
	)

	// This migration only ever runs with a prior version of 23 or below, so the data it loads is always in the
	// pre-25 scheme: the legacy key derivation with no Additional Authenticated Data. The provider is configured for
	// the current scheme, so it's temporarily swapped for the duration.
	key, aad := provider.keys.encryption, provider.aad

	provider.keys.encryption = utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))
	provider.aad = aadNone

	defer func() {
		provider.keys.encryption, provider.aad = key, aad
	}()

	xctx := context.WithValue(ctx, ctxKeyConnection, conn)

	limit := 100

	for page := 0; true; page++ {
		if credentials, err = provider.LoadWebAuthnCredentials(xctx, limit, page); err != nil {
			return fmt.Errorf("failed to verify credentials: %w", err)
		}

		if page == 0 && len(credentials) == 0 {
			return nil
		}

		for _, raw := range credentials {
			attestationType := raw.AttestationType

			if credential, err = raw.ToCredential(); err != nil {
				continue
			}

			if err = credential.VerifyAttestationType(protocol.AttestationPolicy{}, protocol.SignaturePolicy{}); err != nil {
				continue
			}

			if attestationType != credential.AttestationType {
				raw.UpdateAttestationType(credential)

				_ = provider.UpdateWebAuthnCredentialSignIn(xctx, raw)
			}
		}

		if len(credentials) < limit {
			break
		}
	}

	return nil
}

func migrationSpecialDown25(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, before, after, target int) (err error) {
	// The v26 down migration goes straight to the target scheme when the target is below 25, so this stage is
	// skipped when the migration started at 26 or above.
	if prior >= schemaVersionEncryptionAADRowScoped {
		return nil
	}

	encryptKey := utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))

	// When migrating down to a fresh schema all data is destroyed, so there is nothing to re-encrypt into the legacy
	// format. Skipping the re-encryption also avoids a failure when the configured key no longer matches the data (for
	// example after 'storage encryption change-key' without updating the configuration).
	if target != 0 {
		if err = provider.SchemaEncryptionChangeKeyAdvanced(ctx, conn, encryptKey, false, aadColumn, aadNone); err != nil {
			return err
		}
	}

	provider.keys.encryption = encryptKey

	return nil
}

func migrationSpecialUp25(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, before, after, target int) (err error) {
	// The v26 migration re-encrypts everything in a single pass, so when the migration continues past v25 this stage is
	// skipped and v26 carries both the key derivation and the AAD change.
	if target > schemaVersionEncryptionKeyDerivation {
		return nil
	}

	encryptKey := provider.keys.encryption

	if prior != 0 {
		provider.keys.encryption = utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))
	}

	if err = provider.SchemaEncryptionChangeKeyAdvanced(ctx, conn, encryptKey, prior == 0, aadNone, aadColumn); err != nil {
		return err
	}

	provider.keys.encryption = encryptKey

	return nil
}

func migrationSpecialUp26(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, before, after, target int) (err error) {
	encryptKey := provider.keys.encryption

	var decrypt EncryptionAAD

	switch {
	case prior == 0:
		// A fresh schema writes the encryption check value in the current scheme before this runs.
		decrypt = aadRow
	case prior < schemaVersionEncryptionKeyDerivation:
		// The v25 schema is skipped when the target is higher, so this migration also carries the key change.
		provider.keys.encryption = utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))
		decrypt = aadNone
	default:
		decrypt = aadColumn
	}

	if err = provider.SchemaEncryptionChangeKeyAdvanced(ctx, conn, encryptKey, prior == 0, decrypt, aadRow); err != nil {
		return err
	}

	provider.keys.encryption = encryptKey

	return nil
}

func migrationSpecialDown26(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, before, after, target int) (err error) {
	// When migrating down to a fresh schema all data is destroyed, so there is nothing to re-encrypt. The key is
	// still set to the legacy derivation so provider.keys.encryption stays in sync with a schema 0 database, which
	// only ever understood the legacy scheme.
	if target == 0 {
		provider.keys.encryption = utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))

		return nil
	}

	encryptKey := provider.keys.encryption

	encrypt := aadColumn

	if target < schemaVersionEncryptionKeyDerivation {
		encryptKey = utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))
		encrypt = aadNone
	}

	if err = provider.SchemaEncryptionChangeKeyAdvanced(ctx, conn, encryptKey, false, aadRow, encrypt); err != nil {
		return err
	}

	provider.keys.encryption = encryptKey

	return nil
}
