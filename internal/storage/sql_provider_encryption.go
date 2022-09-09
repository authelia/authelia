package storage

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

// SchemaEncryptionChangeKey uses the currently configured key to decrypt values in the database and the key provided
// by this command to encrypt the values again and update them using a transaction.
func (p *SQLProvider) SchemaEncryptionChangeKey(ctx context.Context, encryptionKey string) (err error) {
	tx, err := p.db.Beginx()
	if err != nil {
		return fmt.Errorf("error beginning transaction to change encryption key: %w", err)
	}

	key := sha256.Sum256([]byte(encryptionKey))

	if err = p.schemaEncryptionChangeKeyTOTP(ctx, tx, key); err != nil {
		return err
	}

	if err = p.schemaEncryptionChangeKeyWebauthn(ctx, tx, key); err != nil {
		return err
	}

	if err = p.setNewEncryptionCheckValue(ctx, &key, tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("rollback error %v: rollback due to error: %w", rollbackErr, err)
		}

		return fmt.Errorf("rollback due to error: %w", err)
	}

	return tx.Commit()
}

func (p *SQLProvider) schemaEncryptionChangeKeyTOTP(ctx context.Context, tx *sqlx.Tx, key [32]byte) (err error) {
	var configs []model.TOTPConfiguration

	for page := 0; true; page++ {
		if configs, err = p.LoadTOTPConfigurations(ctx, 10, page); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("rollback error %v: rollback due to error: %w", rollbackErr, err)
			}

			return fmt.Errorf("rollback due to error: %w", err)
		}

		for _, config := range configs {
			if config.Secret, err = utils.Encrypt(config.Secret, &key); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return fmt.Errorf("rollback error %v: rollback due to error: %w", rollbackErr, err)
				}

				return fmt.Errorf("rollback due to error: %w", err)
			}

			if err = p.updateTOTPConfigurationSecret(ctx, config); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return fmt.Errorf("rollback error %v: rollback due to error: %w", rollbackErr, err)
				}

				return fmt.Errorf("rollback due to error: %w", err)
			}
		}

		if len(configs) != 10 {
			break
		}
	}

	return nil
}

func (p *SQLProvider) schemaEncryptionChangeKeyWebauthn(ctx context.Context, tx *sqlx.Tx, key [32]byte) (err error) {
	var devices []model.WebauthnDevice

	for page := 0; true; page++ {
		if devices, err = p.LoadWebauthnDevices(ctx, 10, page); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("rollback error %v: rollback due to error: %w", rollbackErr, err)
			}

			return fmt.Errorf("rollback due to error: %w", err)
		}

		for _, device := range devices {
			if device.PublicKey, err = utils.Encrypt(device.PublicKey, &key); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return fmt.Errorf("rollback error %v: rollback due to error: %w", rollbackErr, err)
				}

				return fmt.Errorf("rollback due to error: %w", err)
			}

			if err = p.updateWebauthnDevicePublicKey(ctx, device); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return fmt.Errorf("rollback error %v: rollback due to error: %w", rollbackErr, err)
				}

				return fmt.Errorf("rollback due to error: %w", err)
			}
		}

		if len(devices) != 10 {
			break
		}
	}

	return nil
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

	if _, err = p.getEncryptionValue(ctx, encryptionNameCheck); err != nil {
		errs = append(errs, ErrSchemaEncryptionInvalidKey)
	}

	if verbose {
		if err = p.schemaEncryptionCheckTOTP(ctx); err != nil {
			errs = append(errs, err)
		}

		if err = p.schemaEncryptionCheckWebauthn(ctx); err != nil {
			errs = append(errs, err)
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

func (p *SQLProvider) schemaEncryptionCheckTOTP(ctx context.Context) (err error) {
	var (
		config  model.TOTPConfiguration
		row     int
		invalid int
		total   int
	)

	pageSize := 10

	var rows *sqlx.Rows

	for page := 0; true; page++ {
		if rows, err = p.db.QueryxContext(ctx, p.sqlSelectTOTPConfigs, pageSize, pageSize*page); err != nil {
			_ = rows.Close()

			return fmt.Errorf("error selecting TOTP configurations: %w", err)
		}

		row = 0

		for rows.Next() {
			total++
			row++

			if err = rows.StructScan(&config); err != nil {
				_ = rows.Close()
				return fmt.Errorf("error scanning TOTP configuration to struct: %w", err)
			}

			if _, err = p.decrypt(config.Secret); err != nil {
				invalid++
			}
		}

		_ = rows.Close()

		if row < pageSize {
			break
		}
	}

	if invalid != 0 {
		return fmt.Errorf("%d of %d total TOTP secrets were invalid", invalid, total)
	}

	return nil
}

func (p *SQLProvider) schemaEncryptionCheckWebauthn(ctx context.Context) (err error) {
	var (
		device  model.WebauthnDevice
		row     int
		invalid int
		total   int
	)

	pageSize := 10

	var rows *sqlx.Rows

	for page := 0; true; page++ {
		if rows, err = p.db.QueryxContext(ctx, p.sqlSelectWebauthnDevices, pageSize, pageSize*page); err != nil {
			_ = rows.Close()

			return fmt.Errorf("error selecting Webauthn devices: %w", err)
		}

		row = 0

		for rows.Next() {
			total++
			row++

			if err = rows.StructScan(&device); err != nil {
				_ = rows.Close()
				return fmt.Errorf("error scanning Webauthn device to struct: %w", err)
			}

			if _, err = p.decrypt(device.PublicKey); err != nil {
				invalid++
			}
		}

		_ = rows.Close()

		if row < pageSize {
			break
		}
	}

	if invalid != 0 {
		return fmt.Errorf("%d of %d total Webauthn devices were invalid", invalid, total)
	}

	return nil
}

func (p *SQLProvider) encrypt(clearText []byte) (cipherText []byte, err error) {
	return utils.Encrypt(clearText, &p.key)
}

func (p *SQLProvider) decrypt(cipherText []byte) (clearText []byte, err error) {
	return utils.Decrypt(cipherText, &p.key)
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
	valueClearText, err := uuid.NewRandom()
	if err != nil {
		return err
	}

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
