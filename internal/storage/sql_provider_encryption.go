package storage

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/authelia/authelia/v4/internal/models"
	"github.com/authelia/authelia/v4/internal/utils"
)

// SchemaEncryptionChangeKey uses the currently configured key to decrypt values in the database and the key provided
// by this command to encrypt the values again and update them using a transaction.
func (p *SQLProvider) SchemaEncryptionChangeKey(ctx context.Context, encryptionKey string) (err error) {
	tx, err := p.db.Beginx()
	if err != nil {
		return err
	}

	key := sha256.Sum256([]byte(encryptionKey))

	for page := 0; true; page++ {
		configs, err := p.LoadTOTPConfigurations(ctx, page, 10)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("%v: %w", rollbackErr, err)
			}

			return err
		}

		for _, config := range configs {
			if config.Secret, err = utils.Encrypt(config.Secret, &key); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return fmt.Errorf("%v: %w", rollbackErr, err)
				}

				return err
			}

			if err = p.UpdateTOTPConfigurationSecret(ctx, config); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return fmt.Errorf("%v: %w", rollbackErr, err)
				}

				return err
			}
		}

		if len(configs) != 10 {
			break
		}
	}

	if err = p.setNewEncryptionCheckValue(ctx, &key, tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("%v: %w", rollbackErr, err)
		}

		return err
	}

	return tx.Commit()
}

// SchemaEncryptionCheckKey checks the encryption key configured is valid for the database.
func (p *SQLProvider) SchemaEncryptionCheckKey(ctx context.Context, verbose bool) (err error) {
	version, err := p.SchemaVersion(ctx)
	if err != nil {
		return err
	}

	if version < 1 {
		return ErrSchemaEncryptionVersionUnsupported
	}

	var errs []error

	_, err = p.getEncryptionValue(ctx, encryptionNameCheck)
	if err != nil {
		errs = append(errs, ErrSchemaEncryptionInvalidKey)
	}

	if verbose {
		var (
			config  models.TOTPConfiguration
			row     int
			invalid int
			total   int
		)

		pageSize := 10

		for page := 0; true; page++ {
			rows, err := p.db.QueryxContext(ctx, p.sqlSelectTOTPConfigs, pageSize, pageSize*page)
			if err != nil {
				_ = rows.Close()

				return err
			}

			row = 0

			for rows.Next() {
				total++
				row++

				err = rows.StructScan(&config)
				if err != nil {
					_ = rows.Close()
					return err
				}

				_, err := p.decrypt(config.Secret)
				if err != nil {
					invalid++
				}
			}

			_ = rows.Close()

			if row < pageSize {
				break
			}
		}

		if invalid != 0 {
			errs = append(errs, fmt.Errorf("%d of %d total totp secrets were invalid", invalid, total))
		}
	}

	if len(errs) != 0 {
		for i, e := range errs {
			if i == 0 {
				err = e

				continue
			}

			err = fmt.Errorf("%w, %v", err, e)
		}

		return err
	}

	return nil
}

func (p SQLProvider) encrypt(clearText []byte) (cipherText []byte, err error) {
	return utils.Encrypt(clearText, p.key)
}

func (p SQLProvider) decrypt(cipherText []byte) (clearText []byte, err error) {
	return utils.Decrypt(cipherText, p.key)
}

func (p *SQLProvider) getEncryptionValue(ctx context.Context, name string) (value []byte, err error) {
	var encryptedValue []byte

	err = p.db.GetContext(ctx, &encryptedValue, p.sqlSelectEncryptionValue, name)
	if err != nil {
		return nil, err
	}

	return p.decrypt(encryptedValue)
}

func (p *SQLProvider) setNewEncryptionCheckValue(ctx context.Context, key *[32]byte, e sqlx.ExecerContext) (err error) {
	valueClearText := uuid.New()

	value, err := utils.Encrypt([]byte(valueClearText.String()), key)
	if err != nil {
		return err
	}

	if e != nil {
		_, err = e.ExecContext(ctx, p.sqlUpsertEncryptionValue, encryptionNameCheck, value)
	} else {
		_, err = p.db.ExecContext(ctx, p.sqlUpsertEncryptionValue, encryptionNameCheck, value)
	}

	return err
}
