package storage

import (
	"context"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

type fSchemaMigration func(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, before, after, target int) (err error)

var migrationsSpecialUp = map[int][]fSchemaMigration{
	24: {migrationSpecialUp24},
	25: {migrationSpecialUp25},
}

var migrationsSpecialDown = map[int][]fSchemaMigration{
	25: {migrationSpecialDown25},
}

func migrationSpecialUp24(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, before, after, target int) (err error) {
	var (
		credentials []model.WebAuthnCredential
		credential  *webauthn.Credential
	)

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

			if err = credential.VerifyAttestationType(); err != nil {
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
	encryptKey := utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))

	// When migrating down to a fresh schema all data is destroyed, so there is nothing to re-encrypt into the legacy
	// format. Skipping the re-encryption also avoids a failure when the configured key no longer matches the data (for
	// example after 'storage encryption change-key' without updating the configuration).
	if target != 0 {
		if err = provider.SchemaEncryptionChangeKeyAdvanced(ctx, conn, encryptKey, false, true, false); err != nil {
			return err
		}
	}

	provider.keys.encryption = encryptKey

	return nil
}

func migrationSpecialUp25(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, before, after, target int) (err error) {
	encryptKey := provider.keys.encryption

	if prior != 0 {
		provider.keys.encryption = utils.DeriveLegacyCryptographicKey([]byte(provider.config.Storage.EncryptionKey))
	}

	if err = provider.SchemaEncryptionChangeKeyAdvanced(ctx, conn, encryptKey, prior == 0, false, true); err != nil {
		return err
	}

	provider.keys.encryption = encryptKey

	return nil
}
