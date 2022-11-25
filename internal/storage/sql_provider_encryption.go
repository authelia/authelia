package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/authelia/authelia/v4/internal/utils"
)

// SchemaEncryptionChangeKey uses the currently configured key to decrypt values in the database and the key provided
// by this command to encrypt the values again and update them using a transaction.
func (p *SQLProvider) SchemaEncryptionChangeKey(ctx context.Context, rawKey string) (err error) {
	key := sha256.Sum256([]byte(rawKey))

	if bytes.Equal(key[:], p.key[:]) {
		return fmt.Errorf("error changing the storage encryption key: the old key and the new key are the same")
	}

	if err = p.SchemaEncryptionCheckKey(ctx, false); err != nil {
		return fmt.Errorf("error changing the storage encryption key: %w", err)
	}

	tx, err := p.db.Beginx()
	if err != nil {
		return fmt.Errorf("error beginning transaction to change encryption key: %w", err)
	}

	encChangeFuncs := []EncryptionChangeKeyFunc{
		schemaEncryptionChangeKeyTOTP,
		schemaEncryptionChangeKeyWebauthn,
	}

	for i := 0; true; i++ {
		typeOAuth2Session := OAuth2SessionType(i)

		if typeOAuth2Session.Table() == "" {
			break
		}

		encChangeFuncs = append(encChangeFuncs, schemaEncryptionChangeKeyOpenIDConnect(typeOAuth2Session))
	}

	for _, encChangeFunc := range encChangeFuncs {
		if err = encChangeFunc(ctx, p, tx, key); err != nil {
			if rerr := tx.Rollback(); rerr != nil {
				return fmt.Errorf("rollback error %v: rollback due to error: %w", rerr, err)
			}

			return fmt.Errorf("rollback due to error: %w", err)
		}
	}

	if err = p.setNewEncryptionCheckValue(ctx, tx, &key); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("rollback error %v: rollback due to error: %w", rollbackErr, err)
		}

		return fmt.Errorf("rollback due to error: %w", err)
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

	if _, err = p.getEncryptionValue(ctx, encryptionNameCheck); err != nil {
		errs = append(errs, ErrSchemaEncryptionInvalidKey)
	}

	if verbose {
		encCheckFuncs := []EncryptionCheckKeyFunc{
			schemaEncryptionCheckKeyTOTP,
			schemaEncryptionCheckKeyWebauthn,
		}

		for i := 0; true; i++ {
			typeOAuth2Session := OAuth2SessionType(i)

			if typeOAuth2Session.Table() == "" {
				break
			}

			encCheckFuncs = append(encCheckFuncs, schemaEncryptionCheckKeyOpenIDConnect(typeOAuth2Session))
		}

		for _, encCheckFunc := range encCheckFuncs {
			if err = encCheckFunc(ctx, p); err != nil {
				errs = append(errs, err)
			}
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

func schemaEncryptionChangeKeyTOTP(ctx context.Context, provider *SQLProvider, tx *sqlx.Tx, key [32]byte) (err error) {
	var count int

	if err = tx.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, tableTOTPConfigurations)); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	configs := make([]encTOTPConfiguration, 0, count)

	if err = tx.SelectContext(ctx, &configs, fmt.Sprintf(queryFmtSelectTOTPConfigurationsEncryptedData, tableTOTPConfigurations)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error selecting TOTP configurations: %w", err)
	}

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateTOTPConfigurationSecret, tableTOTPConfigurations))

	for _, c := range configs {
		if c.Secret, err = provider.decrypt(c.Secret); err != nil {
			return fmt.Errorf("error decrypting TOTP configuration secret with id '%d': %w", c.ID, err)
		}

		if c.Secret, err = utils.Encrypt(c.Secret, &key); err != nil {
			return fmt.Errorf("error encrypting TOTP configuration secret with id '%d': %w", c.ID, err)
		}

		if _, err = tx.ExecContext(ctx, query, c.Secret, c.ID); err != nil {
			return fmt.Errorf("error updating TOTP configuration secret with id '%d': %w", c.ID, err)
		}
	}

	return nil
}

func schemaEncryptionChangeKeyWebauthn(ctx context.Context, provider *SQLProvider, tx *sqlx.Tx, key [32]byte) (err error) {
	var count int

	if err = tx.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, tableWebauthnDevices)); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	devices := make([]encWebauthnDevice, 0, count)

	if err = tx.SelectContext(ctx, &devices, fmt.Sprintf(queryFmtSelectWebauthnDevicesEncryptedData, tableWebauthnDevices)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error selecting Webauthn devices: %w", err)
	}

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateWebauthnDevicePublicKey, tableWebauthnDevices))

	for _, d := range devices {
		if d.PublicKey, err = provider.decrypt(d.PublicKey); err != nil {
			return fmt.Errorf("error decrypting Webauthn device public key with id '%d': %w", d.ID, err)
		}

		if d.PublicKey, err = utils.Encrypt(d.PublicKey, &key); err != nil {
			return fmt.Errorf("error encrypting Webauthn device public key with id '%d': %w", d.ID, err)
		}

		if _, err = tx.ExecContext(ctx, query, d.PublicKey, d.ID); err != nil {
			return fmt.Errorf("error updating Webauthn device public key with id '%d': %w", d.ID, err)
		}
	}

	return nil
}

func schemaEncryptionChangeKeyOpenIDConnect(typeOAuth2Session OAuth2SessionType) EncryptionChangeKeyFunc {
	return func(ctx context.Context, provider *SQLProvider, tx *sqlx.Tx, key [32]byte) (err error) {
		var count int

		if err = tx.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, typeOAuth2Session.Table())); err != nil {
			return err
		}

		if count == 0 {
			return nil
		}

		sessions := make([]encOAuth2Session, 0, count)

		if err = tx.SelectContext(ctx, &sessions, fmt.Sprintf(queryFmtSelectOAuth2SessionEncryptedData, typeOAuth2Session.Table())); err != nil {
			return fmt.Errorf("error selecting oauth2 %s sessions: %w", typeOAuth2Session.String(), err)
		}

		query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateOAuth2ConsentSessionSessionData, typeOAuth2Session.Table()))

		for _, s := range sessions {
			if s.Session, err = provider.decrypt(s.Session); err != nil {
				return fmt.Errorf("error decrypting oauth2 %s session data with id '%d': %w", typeOAuth2Session.String(), s.ID, err)
			}

			if s.Session, err = utils.Encrypt(s.Session, &key); err != nil {
				return fmt.Errorf("error encrypting oauth2 %s session data with id '%d': %w", typeOAuth2Session.String(), s.ID, err)
			}

			if _, err = tx.ExecContext(ctx, query, s.Session, s.ID); err != nil {
				return fmt.Errorf("error updating oauth2 %s session data with id '%d': %w", typeOAuth2Session.String(), s.ID, err)
			}
		}

		return nil
	}
}

func schemaEncryptionCheckKeyTOTP(ctx context.Context, provider *SQLProvider) (err error) {
	var rows *sqlx.Rows

	if rows, err = provider.db.QueryxContext(ctx, fmt.Sprintf(queryFmtSelectTOTPConfigurationsEncryptedData, tableTOTPConfigurations)); err != nil {
		return fmt.Errorf("error selecting TOTP configurations: %w", err)
	}

	var total, invalid int

	var config encTOTPConfiguration

	for rows.Next() {
		total++

		if err = rows.StructScan(&config); err != nil {
			_ = rows.Close()

			return fmt.Errorf("error scanning TOTP configuration to struct: %w", err)
		}

		if _, err = provider.decrypt(config.Secret); err != nil {
			invalid++
		}
	}

	_ = rows.Close()

	if invalid != 0 {
		return fmt.Errorf("%d of %d total TOTP configurations were encrypted with another encryption key", invalid, total)
	}

	return nil
}

func schemaEncryptionCheckKeyWebauthn(ctx context.Context, provider *SQLProvider) (err error) {
	var rows *sqlx.Rows

	if rows, err = provider.db.QueryxContext(ctx, fmt.Sprintf(queryFmtSelectWebauthnDevicesEncryptedData, tableWebauthnDevices)); err != nil {
		return fmt.Errorf("error selecting Webauthn devices: %w", err)
	}

	var total, invalid int

	var device encWebauthnDevice

	for rows.Next() {
		total++

		if err = rows.StructScan(&device); err != nil {
			_ = rows.Close()

			return fmt.Errorf("error scanning Webauthn device to struct: %w", err)
		}

		if _, err = provider.decrypt(device.PublicKey); err != nil {
			invalid++
		}
	}

	_ = rows.Close()

	if invalid != 0 {
		return fmt.Errorf("%d of %d total Webauthn devices were encrypted with another encryption key", invalid, total)
	}

	return nil
}

func schemaEncryptionCheckKeyOpenIDConnect(typeOAuth2Session OAuth2SessionType) EncryptionCheckKeyFunc {
	return func(ctx context.Context, provider *SQLProvider) (err error) {
		var rows *sqlx.Rows

		var total, invalid int

		if rows, err = provider.db.QueryxContext(ctx, fmt.Sprintf(queryFmtSelectOAuth2SessionEncryptedData, typeOAuth2Session.Table())); err != nil {
			return fmt.Errorf("error selecting oauth2 %s sessions: %w", typeOAuth2Session.String(), err)
		}

		var session encOAuth2Session

		for rows.Next() {
			total++

			if err = rows.StructScan(&session); err != nil {
				_ = rows.Close()

				return fmt.Errorf("error scanning oauth2 %s session to struct: %w", typeOAuth2Session.String(), err)
			}

			if _, err = provider.decrypt(session.Session); err != nil {
				invalid++
			}
		}

		_ = rows.Close()

		if invalid != 0 {
			return fmt.Errorf("%d of %d total oauth2 %s session rows were encrypted with another encryption key", invalid, total, typeOAuth2Session.String())
		}

		return nil
	}
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

func (p *SQLProvider) setNewEncryptionCheckValue(ctx context.Context, conn SQLXConnection, key *[32]byte) (err error) {
	valueClearText, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	value, err := utils.Encrypt([]byte(valueClearText.String()), key)
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, p.sqlUpsertEncryptionValue, encryptionNameCheck, value)

	return err
}
