package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/authelia/authelia/v4/internal/utils"
)

// SchemaEncryptionChangeKey uses the currently configured key to decrypt values in the storage provider and the key
// provided by this command to encrypt the values again and update them using a transaction.
func (p *SQLProvider) SchemaEncryptionChangeKey(ctx context.Context, rawKey string) (err error) {
	key := sha256.Sum256([]byte(rawKey))

	if bytes.Equal(key[:], p.keys.encryption[:]) {
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
		schemaEncryptionChangeKeyOneTimeCode,
		schemaEncryptionChangeKeyTOTP,
		schemaEncryptionChangeKeyWebAuthn,
		schemaEncryptionChangeKeyCachedData,
	}

	for i := 0; true; i++ {
		typeOAuth2Session := OAuth2SessionType(i)

		if typeOAuth2Session.Table() == "" {
			break
		}

		encChangeFuncs = append(encChangeFuncs, schemaEncryptionChangeKeyOpenIDConnect(typeOAuth2Session))
	}

	encChangeFuncs = append(encChangeFuncs, schemaEncryptionChangeKeyEncryption)

	for _, encChangeFunc := range encChangeFuncs {
		if err = encChangeFunc(ctx, p, tx, key); err != nil {
			if rerr := tx.Rollback(); rerr != nil {
				return fmt.Errorf("rollback error %v: rollback due to error: %w", rerr, err)
			}

			return fmt.Errorf("rollback due to error: %w", err)
		}
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
			schemaEncryptionCheckKeyOneTimeCode,
			schemaEncryptionCheckKeyTOTP,
			schemaEncryptionCheckKeyWebAuthn,
			schemaEncryptionCheckKeyCachedData,
		}

		for i := 0; true; i++ {
			typeOAuth2Session := OAuth2SessionType(i)

			if typeOAuth2Session.Table() == "" {
				break
			}

			encCheckFuncs = append(encCheckFuncs, schemaEncryptionCheckKeyOpenIDConnect(typeOAuth2Session))
		}

		encCheckFuncs = append(encCheckFuncs, schemaEncryptionCheckKeyEncryption)

		for _, encCheckFunc := range encCheckFuncs {
			table, tableResult := encCheckFunc(ctx, p)

			result.Tables[table] = tableResult
		}
	}

	return result, nil
}

func schemaEncryptionChangeKeyOneTimeCode(ctx context.Context, provider *SQLProvider, tx *sqlx.Tx, key [32]byte) (err error) {
	var count int

	if err = tx.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, tableOneTimeCode)); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	configs := make([]encOneTimeCode, 0, count)

	if err = tx.SelectContext(ctx, &configs, fmt.Sprintf(queryFmtSelectOTCEncryptedData, tableOneTimeCode)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error selecting one-time codes: %w", err)
	}

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateOTCEncryptedData, tableOneTimeCode))

	for _, c := range configs {
		if c.Code, err = provider.decrypt(c.Code); err != nil {
			return fmt.Errorf("error decrypting one-time code with id '%d': %w", c.ID, err)
		}

		if c.Code, err = utils.Encrypt(c.Code, &key); err != nil {
			return fmt.Errorf("error encrypting one-time code with id '%d': %w", c.ID, err)
		}

		if _, err = tx.ExecContext(ctx, query, c.Code, c.ID); err != nil {
			return fmt.Errorf("error updating one-time code with id '%d': %w", c.ID, err)
		}
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

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateTOTPConfigurationEncryptedData, tableTOTPConfigurations))

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

		if d.Attestation != nil {
			if d.Attestation, err = provider.decrypt(d.Attestation); err != nil {
				return fmt.Errorf("error decrypting WebAuthn credential attestation with id '%d': %w", d.ID, err)
			}

			if d.Attestation, err = utils.Encrypt(d.Attestation, &key); err != nil {
				return fmt.Errorf("error encrypting WebAuthn credential attestation with id '%d': %w", d.ID, err)
			}
		}

		if _, err = tx.ExecContext(ctx, query, d.PublicKey, d.Attestation, d.ID); err != nil {
			return fmt.Errorf("error updating WebAuthn credential encrypted columns with id '%d': %w", d.ID, err)
		}
	}

	return nil
}

func schemaEncryptionChangeKeyCachedData(ctx context.Context, provider *SQLProvider, tx *sqlx.Tx, key [32]byte) (err error) {
	var caches []encCachedData

	if err = tx.SelectContext(ctx, &caches, fmt.Sprintf(queryFmtSelectCachedDataEncryptedData, tableCachedData)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error selecting cached data: %w", err)
	}

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateCachedDataEncryptedData, tableCachedData))

	for _, d := range caches {
		if len(d.Value) == 0 {
			continue
		}

		if d.Value, err = provider.decrypt(d.Value); err != nil {
			return fmt.Errorf("error decrypting cached data value id '%d': %w", d.ID, err)
		}

		if d.Value, err = utils.Encrypt(d.Value, &key); err != nil {
			return fmt.Errorf("error encrypting cached data value id '%d': %w", d.ID, err)
		}

		if _, err = tx.ExecContext(ctx, query, d.Value, d.ID); err != nil {
			return fmt.Errorf("error updating cached data encrypted columns with id '%d': %w", d.ID, err)
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

		query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateOAuth2ConsentSessionEncryptedData, typeOAuth2Session.Table()))

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

func schemaEncryptionChangeKeyEncryption(ctx context.Context, provider *SQLProvider, tx *sqlx.Tx, key [32]byte) (err error) {
	var count int

	if err = tx.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, tableEncryption)); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	configs := make([]encEncryption, 0, count)

	if err = tx.SelectContext(ctx, &configs, fmt.Sprintf(queryFmtSelectEncryptionEncryptedData, tableEncryption)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error selecting encyption value: %w", err)
	}

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateEncryptionEncryptedData, tableEncryption))

	for _, c := range configs {
		if c.Value, err = provider.decrypt(c.Value); err != nil {
			return fmt.Errorf("error decrypting encyption value with id '%d': %w", c.ID, err)
		}

		if c.Value, err = utils.Encrypt(c.Value, &key); err != nil {
			return fmt.Errorf("error encrypting encyption value with id '%d': %w", c.ID, err)
		}

		if _, err = tx.ExecContext(ctx, query, c.Value, c.ID); err != nil {
			return fmt.Errorf("error updating encyption value with id '%d': %w", c.ID, err)
		}
	}

	return nil
}

func schemaEncryptionCheckKeyOneTimeCode(ctx context.Context, provider *SQLProvider) (table string, result EncryptionValidationTableResult) {
	var (
		rows *sqlx.Rows
		err  error
	)

	if rows, err = provider.db.QueryxContext(ctx, fmt.Sprintf(queryFmtSelectOTCEncryptedData, tableOneTimeCode)); err != nil {
		return tableOneTimeCode, EncryptionValidationTableResult{Error: fmt.Errorf("error selecting one time-codes: %w", err)}
	}

	var config encOneTimeCode

	for rows.Next() {
		result.Total++

		if err = rows.StructScan(&config); err != nil {
			_ = rows.Close()

			return tableOneTimeCode, EncryptionValidationTableResult{Error: fmt.Errorf("error scanning one time-code to struct: %w", err)}
		}

		if _, err = provider.decrypt(config.Code); err != nil {
			result.Invalid++
		}
	}

	_ = rows.Close()

	return tableOneTimeCode, result
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
		} else if credential.Attestation != nil {
			if _, err = provider.decrypt(credential.Attestation); err != nil {
				result.Invalid++
			}
		}
	}

	_ = rows.Close()

	return tableWebAuthnCredentials, result
}

func schemaEncryptionCheckKeyCachedData(ctx context.Context, provider *SQLProvider) (table string, result EncryptionValidationTableResult) {
	var (
		rows *sqlx.Rows
		err  error
	)

	if rows, err = provider.db.QueryxContext(ctx, fmt.Sprintf(queryFmtSelectCachedDataEncryptedData, tableCachedData)); err != nil {
		return tableCachedData, EncryptionValidationTableResult{Error: fmt.Errorf("error selecting cached data: %w", err)}
	}

	var cache encCachedData

	for rows.Next() {
		result.Total++

		if err = rows.StructScan(&cache); err != nil {
			_ = rows.Close()

			return tableCachedData, EncryptionValidationTableResult{Error: fmt.Errorf("error scanning cached data to struct: %w", err)}
		}

		if _, err = provider.decrypt(cache.Value); err != nil {
			result.Invalid++
		}
	}

	_ = rows.Close()

	return tableCachedData, result
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

func schemaEncryptionCheckKeyEncryption(ctx context.Context, provider *SQLProvider) (table string, result EncryptionValidationTableResult) {
	var (
		rows *sqlx.Rows
		err  error
	)

	if rows, err = provider.db.QueryxContext(ctx, fmt.Sprintf(queryFmtSelectEncryptionEncryptedData, tableEncryption)); err != nil {
		return tableEncryption, EncryptionValidationTableResult{Error: fmt.Errorf("error selecting encryption values: %w", err)}
	}

	var config encEncryption

	for rows.Next() {
		result.Total++

		if err = rows.StructScan(&config); err != nil {
			_ = rows.Close()

			return tableEncryption, EncryptionValidationTableResult{Error: fmt.Errorf("error scanning encryption value to struct: %w", err)}
		}

		if _, err = provider.decrypt(config.Value); err != nil {
			result.Invalid++
		}
	}

	_ = rows.Close()

	return tableEncryption, result
}

func (p *SQLProvider) encrypt(clearText []byte) (cipherText []byte, err error) {
	return utils.Encrypt(clearText, &p.keys.encryption)
}

func (p *SQLProvider) decrypt(cipherText []byte) (clearText []byte, err error) {
	return utils.Decrypt(cipherText, &p.keys.encryption)
}

func (p *SQLProvider) otcHMACSignature(values ...[]byte) string {
	h := hmac.New(sha512.New, p.keys.otcHMAC)

	for i := 0; i < len(values); i++ {
		h.Write(values[i])
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (p *SQLProvider) otpHMACSignature(values ...[]byte) string {
	h := hmac.New(sha256.New, p.keys.otpHMAC)

	for i := 0; i < len(values); i++ {
		h.Write(values[i])
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (p *SQLProvider) getHMACOneTimeCode(ctx context.Context) (key []byte, err error) {
	return p.getHMACKey(ctx, "hmac_key_otc", sha512.BlockSize)
}

func (p *SQLProvider) getHMACOneTimePassword(ctx context.Context) (key []byte, err error) {
	return p.getHMACKey(ctx, "hmac_key_otp", sha256.BlockSize)
}

func (p *SQLProvider) getHMACKey(ctx context.Context, name string, size int) (key []byte, err error) {
	if key, err = p.getEncryptionValue(ctx, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			key = make([]byte, size)

			_, err = rand.Read(key)

			if err != nil {
				return nil, fmt.Errorf("failed to generate hmac key: %w", err)
			}

			if err = p.setEncryptionValue(ctx, name, key); err != nil {
				return nil, err
			}

			return key, nil
		}

		return nil, err
	}

	return key, nil
}

func (p *SQLProvider) getEncryptionValue(ctx context.Context, name string) (value []byte, err error) {
	var encryptedValue []byte

	err = p.db.GetContext(ctx, &encryptedValue, p.sqlSelectEncryptionValue, name)
	if err != nil {
		return nil, err
	}

	return p.decrypt(encryptedValue)
}

func (p *SQLProvider) setEncryptionValue(ctx context.Context, name string, value []byte) (err error) {
	if value, err = p.encrypt(value); err != nil {
		return err
	}

	if _, err = p.db.ExecContext(ctx, p.sqlUpsertEncryptionValue, name, value); err != nil {
		return err
	}

	return nil
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
