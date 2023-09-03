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
func (p *SQLProvider) SchemaEncryptionChangeKey(ctx context.Context, key string) (err error) {
	skey := sha256.Sum256([]byte(key))

	if bytes.Equal(skey[:], p.key[:]) {
		return fmt.Errorf("error changing the storage encryption key: the old key and the new key are the same")
	}

	if _, err = p.SchemaEncryptionCheckKey(ctx, false); err != nil {
		return fmt.Errorf("error changing the storage encryption key: %w", err)
	}

	tx, err := p.db.Beginx()
	if err != nil {
		return fmt.Errorf("error beginning transaction to change encryption key: %w", err)
	}

	encChangeFuncs := []EncryptionChangeKeyFunc{
		schemaEncryptionChangeKeyTOTP,
		schemaEncryptionChangeKeyWebAuthn,
	}

	for i := 0; true; i++ {
		typeOAuth2Session := OAuth2SessionType(i)

		if typeOAuth2Session.Table() == "" {
			break
		}

		encChangeFuncs = append(encChangeFuncs, schemaEncryptionChangeKeyOpenIDConnect(typeOAuth2Session))
	}

	for _, encChangeFunc := range encChangeFuncs {
		if err = encChangeFunc(ctx, p, tx, skey); err != nil {
			if rerr := tx.Rollback(); rerr != nil {
				return fmt.Errorf("rollback error %v: rollback due to error: %w", rerr, err)
			}

			return fmt.Errorf("rollback due to error: %w", err)
		}
	}

	if err = p.setNewEncryptionCheckValue(ctx, tx, &skey); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("rollback error %v: rollback due to error: %w", rerr, err)
		}

		return fmt.Errorf("rollback due to error: %w", err)
	}

	return tx.Commit()
}

// SchemaEncryptionCheckKey checks the encryption key configured is valid for the database.
func (p *SQLProvider) SchemaEncryptionCheckKey(ctx context.Context, verbose bool) (result EncryptionValidationResult, err error) {
	version, err := p.SchemaVersion(ctx)
	if err != nil {
		return result, err
	}

	if version < 1 {
		return result, ErrSchemaEncryptionVersionUnsupported
	}

	result = EncryptionValidationResult{
		Tables: map[string]EncryptionValidationTableResult{},
	}

	if _, err = p.getEncryptionValue(ctx, encryptionNameCheck); err != nil {
		result.InvalidCheckValue = true
	}

	if verbose {
		encCheckFuncs := []EncryptionCheckKeyFunc{
			schemaEncryptionCheckKeyTOTP,
			schemaEncryptionCheckKeyWebAuthn,
		}

		for i := 0; true; i++ {
			typeOAuth2Session := OAuth2SessionType(i)

			if typeOAuth2Session.Table() == "" {
				break
			}

			encCheckFuncs = append(encCheckFuncs, schemaEncryptionCheckKeyOpenIDConnect(typeOAuth2Session))
		}

		for _, encCheckFunc := range encCheckFuncs {
			table, tableResult := encCheckFunc(ctx, p)

			result.Tables[table] = tableResult
		}
	}

	return result, nil
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

func schemaEncryptionChangeKeyWebAuthn(ctx context.Context, provider *SQLProvider, tx *sqlx.Tx, key [32]byte) (err error) {
	var count int

	if err = tx.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, tableWebAuthnCredentials)); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	credentials := make([]encWebAuthnCredential, 0, count)

	if err = tx.SelectContext(ctx, &credentials, fmt.Sprintf(queryFmtSelectWebAuthnCredentialsEncryptedData, tableWebAuthnCredentials)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error selecting WebAuthn credentials: %w", err)
	}

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateWebAuthnCredentialsEncryptedData, tableWebAuthnCredentials))

	for _, d := range credentials {
		if d.PublicKey, err = provider.decrypt(d.PublicKey); err != nil {
			return fmt.Errorf("error decrypting WebAuthn credential public key with id '%d': %w", d.ID, err)
		}

		if d.PublicKey, err = utils.Encrypt(d.PublicKey, &key); err != nil {
			return fmt.Errorf("error encrypting WebAuthn credential public key with id '%d': %w", d.ID, err)
		}

		if _, err = tx.ExecContext(ctx, query, d.PublicKey, d.ID); err != nil {
			return fmt.Errorf("error updating WebAuthn credential public key with id '%d': %w", d.ID, err)
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

func schemaEncryptionCheckKeyTOTP(ctx context.Context, provider *SQLProvider) (table string, result EncryptionValidationTableResult) {
	var (
		rows *sqlx.Rows
		err  error
	)

	if rows, err = provider.db.QueryxContext(ctx, fmt.Sprintf(queryFmtSelectTOTPConfigurationsEncryptedData, tableTOTPConfigurations)); err != nil {
		return tableTOTPConfigurations, EncryptionValidationTableResult{Error: fmt.Errorf("error selecting TOTP configurations: %w", err)}
	}

	var config encTOTPConfiguration

	for rows.Next() {
		result.Total++

		if err = rows.StructScan(&config); err != nil {
			_ = rows.Close()

			return tableTOTPConfigurations, EncryptionValidationTableResult{Error: fmt.Errorf("error scanning TOTP configuration to struct: %w", err)}
		}

		if _, err = provider.decrypt(config.Secret); err != nil {
			result.Invalid++
		}
	}

	_ = rows.Close()

	return tableTOTPConfigurations, result
}

func schemaEncryptionCheckKeyWebAuthn(ctx context.Context, provider *SQLProvider) (table string, result EncryptionValidationTableResult) {
	var (
		rows *sqlx.Rows
		err  error
	)

	if rows, err = provider.db.QueryxContext(ctx, fmt.Sprintf(queryFmtSelectWebAuthnCredentialsEncryptedData, tableWebAuthnCredentials)); err != nil {
		return tableWebAuthnCredentials, EncryptionValidationTableResult{Error: fmt.Errorf("error selecting WebAuthn credentials: %w", err)}
	}

	var credential encWebAuthnCredential

	for rows.Next() {
		result.Total++

		if err = rows.StructScan(&credential); err != nil {
			_ = rows.Close()

			return tableWebAuthnCredentials, EncryptionValidationTableResult{Error: fmt.Errorf("error scanning WebAuthn credential to struct: %w", err)}
		}

		if _, err = provider.decrypt(credential.PublicKey); err != nil {
			result.Invalid++
		}
	}

	_ = rows.Close()

	return tableWebAuthnCredentials, result
}

func schemaEncryptionCheckKeyOpenIDConnect(typeOAuth2Session OAuth2SessionType) EncryptionCheckKeyFunc {
	return func(ctx context.Context, provider *SQLProvider) (table string, result EncryptionValidationTableResult) {
		var (
			rows *sqlx.Rows
			err  error
		)

		if rows, err = provider.db.QueryxContext(ctx, fmt.Sprintf(queryFmtSelectOAuth2SessionEncryptedData, typeOAuth2Session.Table())); err != nil {
			return typeOAuth2Session.Table(), EncryptionValidationTableResult{Error: fmt.Errorf("error selecting oauth2 %s sessions: %w", typeOAuth2Session.String(), err)}
		}

		var session encOAuth2Session

		for rows.Next() {
			result.Total++

			if err = rows.StructScan(&session); err != nil {
				_ = rows.Close()

				return typeOAuth2Session.Table(), EncryptionValidationTableResult{Error: fmt.Errorf("error scanning oauth2 %s session to struct: %w", typeOAuth2Session.String(), err)}
			}

			if _, err = provider.decrypt(session.Session); err != nil {
				result.Invalid++
			}
		}

		_ = rows.Close()

		return typeOAuth2Session.Table(), result
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
