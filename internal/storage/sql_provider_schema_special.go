package storage

import (
	"context"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/model"
)

type fSchemaMigration func(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, target int) (err error)

var migrationsSpecialUp = map[int][]fSchemaMigration{
	24: {migrationSpecialUp24},
}

var migrationsSpecialDown = map[int][]fSchemaMigration{}

func migrationSpecialUp24(ctx context.Context, conn SQLXConnection, provider *SQLProvider, prior, target int) (err error) {
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
