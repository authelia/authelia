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
	"github.com/rpadovani/sqlx-v2"

	"github.com/authelia/authelia/v4/internal/utils"
)

func (p *SQLProvider) SchemaEncryptionRotateHMACKey(ctx context.Context, name string) (err error) {
	var (
		size  int
		table string
		desc  string
	)

	switch name {
	case hmacNameOneTimeCode:
		size, table, desc = sha512.BlockSize, tableOneTimeCode, "one time-codes"
	case hmacNameOneTimePassword:
		size, table, desc = sha256.BlockSize, tableTOTPHistory, "totp history"
	default:
		return fmt.Errorf("unknown key name '%s'", name)
	}

	var tx SQLXTx

	if tx, err = p.db.Beginx(); err != nil {
		return fmt.Errorf("error beginning transaction to rotate hmac key: %w", err)
	}

	if _, err = p.setCrypographyKey(ctx, tx, keyTypeCryptographyHMAC, name, size, true); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("error rolling back transaction to rotate hmac key: %w", rollbackErr)
		}

		return fmt.Errorf("error setting the hmac key: %w", err)
	}

	if err = p.truncate(ctx, tx, table); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("error rolling back transaction to rotate hmac key: %w", rollbackErr)
		}

		return fmt.Errorf("error truncating %s: %w", desc, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction to rotate hmac key: %w", err)
	}

	return nil
}

// SchemaEncryptionChangeKey uses the currently configured key to decrypt values in the storage provider and the key
// provided by this command to encrypt the values again and update them using a transaction.
func (p *SQLProvider) SchemaEncryptionChangeKey(ctx context.Context, rawKey string) (err error) {
	var key []byte

	if key, err = utils.DeriveCryptographicKey([]byte(rawKey), hkdfKeyInfo, sha256.New); err != nil {
		return err
	}

	if bytes.Equal(key, p.keys.encryption) {
		return fmt.Errorf("error changing the storage encryption key: the old key and the new key are the same")
	}

	if _, err = p.SchemaEncryptionCheckKey(ctx, false); err != nil {
		return fmt.Errorf("error changing the storage encryption key: %w", err)
	}

	var version int

	if version, err = p.SchemaVersion(ctx); err != nil {
		return fmt.Errorf("error changing the storage encryption key: %w", err)
	}

	aad := aadForSchemaVersion(version)

	var tx SQLXTx

	if tx, err = p.db.Beginx(); err != nil {
		return fmt.Errorf("error beginning transaction to change encryption key: %w", err)
	}

	if err = p.SchemaEncryptionChangeKeyAdvanced(ctx, tx, key, false, aad, aad); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("error rolling back transaction to change encryption key: %w", rollbackErr)
		}

		return fmt.Errorf("error changing the storage encryption key: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction to change encryption key: %w", err)
	}

	return nil
}

func (p *SQLProvider) SchemaEncryptionChangeKeyAdvanced(ctx context.Context, conn SQLXConnection, key []byte, init bool, decrypt, encrypt EncryptionAAD) (err error) {
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
		if err = encChangeFunc(ctx, p, conn, init, decrypt, encrypt, key); err != nil {
			return err
		}
	}

	return nil
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

	if err = p.checkEncryptionCheckValue(ctx, version); err != nil {
		result.InvalidCheckValue = true
	}

	aad := aadForSchemaVersion(version)

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
			table, tableResult := encCheckFunc(ctx, p, aad)

			result.Tables[table] = tableResult
		}
	}

	return result, nil
}

func schemaEncryptionChangeKeyOneTimeCode(ctx context.Context, provider *SQLProvider, conn SQLXConnection, init bool, decrypt, encrypt EncryptionAAD, key []byte) (err error) {
	var count int

	if err = conn.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, tableOneTimeCode)); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	configs := make([]encOneTimeCode, 0, count)

	if err = conn.SelectContext(ctx, &configs, fmt.Sprintf(queryFmtSelectOTCEncryptedData, tableOneTimeCode)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error selecting one-time codes: %w", err)
	}

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateOTCEncryptedData, tableOneTimeCode))

	for _, c := range configs {
		if c.Code, err = utils.Decrypt(c.Code, decrypt.Get(tableOneTimeCode, columnCode, c.Signature), provider.keys.encryption); err != nil {
			return fmt.Errorf("error decrypting one-time code with id '%d': %w", c.ID, err)
		}

		if c.Code, err = utils.Encrypt(c.Code, encrypt.Get(tableOneTimeCode, columnCode, c.Signature), key); err != nil {
			return fmt.Errorf("error encrypting one-time code with id '%d': %w", c.ID, err)
		}

		if _, err = conn.ExecContext(ctx, query, c.Code, c.ID); err != nil {
			return fmt.Errorf("error updating one-time code with id '%d': %w", c.ID, err)
		}
	}

	return nil
}

func schemaEncryptionChangeKeyTOTP(ctx context.Context, provider *SQLProvider, conn SQLXConnection, init bool, decrypt, encrypt EncryptionAAD, key []byte) (err error) {
	var count int

	if err = conn.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, tableTOTPConfigurations)); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	configs := make([]encTOTPConfiguration, 0, count)

	if err = conn.SelectContext(ctx, &configs, fmt.Sprintf(queryFmtSelectTOTPConfigurationsEncryptedData, tableTOTPConfigurations)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error selecting TOTP configurations: %w", err)
	}

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateTOTPConfigurationEncryptedData, tableTOTPConfigurations))

	for _, c := range configs {
		if c.Secret, err = utils.Decrypt(c.Secret, decrypt.Get(tableTOTPConfigurations, columnSecret, c.Username), provider.keys.encryption); err != nil {
			return fmt.Errorf("error decrypting TOTP configuration secret with id '%d': %w", c.ID, err)
		}

		if c.Secret, err = utils.Encrypt(c.Secret, encrypt.Get(tableTOTPConfigurations, columnSecret, c.Username), key); err != nil {
			return fmt.Errorf("error encrypting TOTP configuration secret with id '%d': %w", c.ID, err)
		}

		if _, err = conn.ExecContext(ctx, query, c.Secret, c.ID); err != nil {
			return fmt.Errorf("error updating TOTP configuration secret with id '%d': %w", c.ID, err)
		}
	}

	return nil
}

func schemaEncryptionChangeKeyWebAuthn(ctx context.Context, provider *SQLProvider, conn SQLXConnection, init bool, decrypt, encrypt EncryptionAAD, key []byte) (err error) {
	var count int

	if err = conn.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, tableWebAuthnCredentials)); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	credentials := make([]encWebAuthnCredential, 0, count)

	if err = conn.SelectContext(ctx, &credentials, fmt.Sprintf(queryFmtSelectWebAuthnCredentialsEncryptedData, tableWebAuthnCredentials)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error selecting WebAuthn credentials: %w", err)
	}

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateWebAuthnCredentialsEncryptedData, tableWebAuthnCredentials))

	for _, d := range credentials {
		if d.PublicKey, err = utils.Decrypt(d.PublicKey, decrypt.GetIssuer(tableWebAuthnCredentials, "public_key", d.KID, d.RPID), provider.keys.encryption); err != nil {
			return fmt.Errorf("error decrypting WebAuthn credential public key with id '%d': %w", d.ID, err)
		}

		if d.PublicKey, err = utils.Encrypt(d.PublicKey, encrypt.GetIssuer(tableWebAuthnCredentials, "public_key", d.KID, d.RPID), key); err != nil {
			return fmt.Errorf("error encrypting WebAuthn credential public key with id '%d': %w", d.ID, err)
		}

		if d.Attestation != nil {
			if d.Attestation, err = utils.Decrypt(d.Attestation, decrypt.GetIssuer(tableWebAuthnCredentials, "attestation", d.KID, d.RPID), provider.keys.encryption); err != nil {
				return fmt.Errorf("error decrypting WebAuthn credential attestation with id '%d': %w", d.ID, err)
			}

			if d.Attestation, err = utils.Encrypt(d.Attestation, encrypt.GetIssuer(tableWebAuthnCredentials, "attestation", d.KID, d.RPID), key); err != nil {
				return fmt.Errorf("error encrypting WebAuthn credential attestation with id '%d': %w", d.ID, err)
			}
		}

		if _, err = conn.ExecContext(ctx, query, d.PublicKey, d.Attestation, d.ID); err != nil {
			return fmt.Errorf("error updating WebAuthn credential encrypted columns with id '%d': %w", d.ID, err)
		}
	}

	return nil
}

func schemaEncryptionChangeKeyCachedData(ctx context.Context, provider *SQLProvider, conn SQLXConnection, init bool, decrypt, encrypt EncryptionAAD, key []byte) (err error) {
	var caches []encCachedData

	if err = conn.SelectContext(ctx, &caches, conn.Rebind(fmt.Sprintf(queryFmtSelectCachedDataValueEncrypted, tableCachedData)), true); err != nil {
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

		if d.Value, err = utils.Decrypt(d.Value, decrypt.Get(tableCachedData, columnValue, d.Name), provider.keys.encryption); err != nil {
			return fmt.Errorf("error decrypting cached data value id '%d': %w", d.ID, err)
		}

		if d.Value, err = utils.Encrypt(d.Value, encrypt.Get(tableCachedData, columnValue, d.Name), key); err != nil {
			return fmt.Errorf("error encrypting cached data value id '%d': %w", d.ID, err)
		}

		if _, err = conn.ExecContext(ctx, query, d.Value, d.ID); err != nil {
			return fmt.Errorf("error updating cached data encrypted columns with id '%d': %w", d.ID, err)
		}
	}

	return nil
}

func schemaEncryptionChangeKeyOpenIDConnect(typeOAuth2Session OAuth2SessionType) EncryptionChangeKeyFunc {
	return func(ctx context.Context, provider *SQLProvider, conn SQLXConnection, init bool, decrypt, encrypt EncryptionAAD, key []byte) (err error) {
		var count int

		if err = conn.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, typeOAuth2Session.Table())); err != nil {
			return err
		}

		if count == 0 {
			return nil
		}

		sessions := make([]encOAuth2Session, 0, count)

		if err = conn.SelectContext(ctx, &sessions, fmt.Sprintf(queryFmtSelectOAuth2SessionEncryptedData, typeOAuth2Session.Table())); err != nil {
			return fmt.Errorf("error selecting oauth2 %s sessions: %w", typeOAuth2Session.String(), err)
		}

		query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateOAuth2ConsentSessionEncryptedData, typeOAuth2Session.Table()))

		for _, s := range sessions {
			if s.Session, err = utils.Decrypt(s.Session, decrypt.Get(typeOAuth2Session.AAD(), columnSessionData, s.Signature), provider.keys.encryption); err != nil {
				return fmt.Errorf("error decrypting oauth2 %s session data with id '%d': %w", typeOAuth2Session.String(), s.ID, err)
			}

			if s.Session, err = utils.Encrypt(s.Session, encrypt.Get(typeOAuth2Session.AAD(), columnSessionData, s.Signature), key); err != nil {
				return fmt.Errorf("error encrypting oauth2 %s session data with id '%d': %w", typeOAuth2Session.String(), s.ID, err)
			}

			if _, err = conn.ExecContext(ctx, query, s.Session, s.ID); err != nil {
				return fmt.Errorf("error updating oauth2 %s session data with id '%d': %w", typeOAuth2Session.String(), s.ID, err)
			}
		}

		return nil
	}
}

func schemaEncryptionChangeKeyEncryption(ctx context.Context, provider *SQLProvider, conn SQLXConnection, init bool, decrypt, encrypt EncryptionAAD, key []byte) (err error) {
	var count int

	if err = conn.GetContext(ctx, &count, fmt.Sprintf(queryFmtSelectRowCount, tableEncryption)); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	configs := make([]encEncryption, 0, count)

	if err = conn.SelectContext(ctx, &configs, fmt.Sprintf(queryFmtSelectEncryptionEncryptedData, tableEncryption)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error selecting encryption value: %w", err)
	}

	query := provider.db.Rebind(fmt.Sprintf(queryFmtUpdateEncryptionEncryptedData, tableEncryption))

	decryptAAD := decrypt

	if init {
		decryptAAD = encrypt
	}

	for _, c := range configs {
		if c.Value, err = utils.Decrypt(c.Value, decryptAAD.Get(tableEncryption, columnValue, c.Name), provider.keys.encryption); err != nil {
			return fmt.Errorf("error decrypting encryption value with id '%d': %w", c.ID, err)
		}

		if c.Value, err = utils.Encrypt(c.Value, encrypt.Get(tableEncryption, columnValue, c.Name), key); err != nil {
			return fmt.Errorf("error encrypting encryption value with id '%d': %w", c.ID, err)
		}

		if _, err = conn.ExecContext(ctx, query, c.Value, c.ID); err != nil {
			return fmt.Errorf("error updating encryption value with id '%d': %w", c.ID, err)
		}
	}

	return nil
}

func schemaEncryptionCheckKeyOneTimeCode(ctx context.Context, provider *SQLProvider, aad EncryptionAAD) (table string, result EncryptionValidationTableResult) {
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

		if _, err = utils.Decrypt(config.Code, aad.Get(tableOneTimeCode, columnCode, config.Signature), provider.keys.encryption); err != nil {
			result.Invalid++
		}
	}

	_ = rows.Close()

	return tableOneTimeCode, result
}

func schemaEncryptionCheckKeyTOTP(ctx context.Context, provider *SQLProvider, aad EncryptionAAD) (table string, result EncryptionValidationTableResult) {
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

		if _, err = utils.Decrypt(config.Secret, aad.Get(tableTOTPConfigurations, columnSecret, config.Username), provider.keys.encryption); err != nil {
			result.Invalid++
		}
	}

	_ = rows.Close()

	return tableTOTPConfigurations, result
}

func schemaEncryptionCheckKeyWebAuthn(ctx context.Context, provider *SQLProvider, aad EncryptionAAD) (table string, result EncryptionValidationTableResult) {
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

		if _, err = utils.Decrypt(credential.PublicKey, aad.GetIssuer(tableWebAuthnCredentials, "public_key", credential.KID, credential.RPID), provider.keys.encryption); err != nil {
			result.Invalid++
		} else if credential.Attestation != nil {
			if _, err = utils.Decrypt(credential.Attestation, aad.GetIssuer(tableWebAuthnCredentials, "attestation", credential.KID, credential.RPID), provider.keys.encryption); err != nil {
				result.Invalid++
			}
		}
	}

	_ = rows.Close()

	return tableWebAuthnCredentials, result
}

func schemaEncryptionCheckKeyCachedData(ctx context.Context, provider *SQLProvider, aad EncryptionAAD) (table string, result EncryptionValidationTableResult) {
	var (
		rows *sqlx.Rows
		err  error
	)
	if rows, err = provider.db.QueryxContext(ctx, provider.db.Rebind(fmt.Sprintf(queryFmtSelectCachedDataValueEncrypted, tableCachedData)), true); err != nil {
		return tableCachedData, EncryptionValidationTableResult{Error: fmt.Errorf("error selecting cached data: %w", err)}
	}

	var cache encCachedData

	for rows.Next() {
		result.Total++

		if err = rows.StructScan(&cache); err != nil {
			_ = rows.Close()

			return tableCachedData, EncryptionValidationTableResult{Error: fmt.Errorf("error scanning cached data to struct: %w", err)}
		}

		if _, err = utils.Decrypt(cache.Value, aad.Get(tableCachedData, columnValue, cache.Name), provider.keys.encryption); err != nil {
			result.Invalid++
		}
	}

	_ = rows.Close()

	return tableCachedData, result
}

func schemaEncryptionCheckKeyOpenIDConnect(typeOAuth2Session OAuth2SessionType) EncryptionCheckKeyFunc {
	return func(ctx context.Context, provider *SQLProvider, aad EncryptionAAD) (table string, result EncryptionValidationTableResult) {
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

			if _, err = utils.Decrypt(session.Session, aad.Get(typeOAuth2Session.AAD(), columnSessionData, session.Signature), provider.keys.encryption); err != nil {
				result.Invalid++
			}
		}

		_ = rows.Close()

		return typeOAuth2Session.Table(), result
	}
}

func schemaEncryptionCheckKeyEncryption(ctx context.Context, provider *SQLProvider, aad EncryptionAAD) (table string, result EncryptionValidationTableResult) {
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

		if _, err = utils.Decrypt(config.Value, aad.Get(tableEncryption, columnValue, config.Name), provider.keys.encryption); err != nil {
			result.Invalid++
		}
	}

	_ = rows.Close()

	return tableEncryption, result
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
	return p.getHMACKey(ctx, hmacNameOneTimeCode, sha512.BlockSize)
}

func (p *SQLProvider) getHMACOneTimePassword(ctx context.Context) (key []byte, err error) {
	return p.getHMACKey(ctx, hmacNameOneTimePassword, sha256.BlockSize)
}

func (p *SQLProvider) setCrypographyKey(ctx context.Context, conn SQLXConnection, typ string, name string, size int, replace bool) (key []byte, err error) {
	key = make([]byte, size)

	if _, err = rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	var encName string

	switch typ {
	case keyTypeCryptographyHMAC:
		encName = fmt.Sprintf(fmtNameKeyHMAC, name)
	case keyTypeCryptographyEnc:
		encName = fmt.Sprintf(fmtNameKeyEnc, name)
	default:
		return nil, fmt.Errorf("invalid key type: %s", typ)
	}

	if err = p.setEncryptionValue(ctx, conn, encName, key, replace); err != nil {
		return nil, err
	}

	return key, nil
}

func (p *SQLProvider) getHMACKey(ctx context.Context, name string, size int) (key []byte, err error) {
	var tx SQLXTx

	if tx, err = p.db.BeginTxx(ctx, nil); err != nil {
		return nil, fmt.Errorf("error beginning transaction to get hmac key: %w", err)
	}

	if key, err = p.getEncryptionValue(ctx, tx, fmt.Sprintf(fmtNameKeyHMAC, name)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if key, err = p.setCrypographyKey(ctx, tx, keyTypeCryptographyHMAC, name, size, false); err != nil {
				_ = tx.Rollback()

				return nil, err
			}

			if txerr := tx.Commit(); txerr != nil {
				return nil, fmt.Errorf("error occurred committing transaction to get hmac key: %w", txerr)
			}

			return key, nil
		}

		_ = tx.Rollback()

		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error occurred committing transaction to get hmac key: %w", err)
	}

	return key, nil
}

// checkEncryptionCheckValue reads and decrypts the encryption check value using the key and AAD appropriate for the
// provided schema version, resolved via aadForSchemaVersion. Databases created before HKDF key derivation and GCM
// AAD were introduced (schemaVersionEncryptionKeyDerivation) store the value using the legacy SHA256 key without
// AAD, so validating them with the derived key would incorrectly fail prior to the upgrade migration running.
func (p *SQLProvider) checkEncryptionCheckValue(ctx context.Context, version int) (err error) {
	key, aad := p.keys.encryption, aadForSchemaVersion(version).Get(tableEncryption, columnValue, encryptionNameCheck)

	if version < schemaVersionEncryptionKeyDerivation {
		key = utils.DeriveLegacyCryptographicKey([]byte(p.config.Storage.EncryptionKey))
	}

	var encryptedValue []byte

	if err = p.db.GetContext(ctx, &encryptedValue, p.sqlSelectEncryptionValue, encryptionNameCheck); err != nil {
		return err
	}

	if _, err = utils.Decrypt(encryptedValue, aad, key); err != nil {
		return err
	}

	return nil
}

func (p *SQLProvider) getEncryptionValue(ctx context.Context, conn SQLXConnection, name string) (value []byte, err error) {
	var encryptedValue []byte

	if err = conn.GetContext(ctx, &encryptedValue, p.sqlSelectEncryptionValue, name); err != nil {
		return nil, err
	}

	if value, err = utils.Decrypt(encryptedValue, p.aad.Get(tableEncryption, columnValue, name), p.keys.encryption); err != nil {
		return nil, err
	}

	return value, nil
}

func (p *SQLProvider) setEncryptionValue(ctx context.Context, conn SQLXConnection, name string, value []byte, replace bool) (err error) {
	if value, err = utils.Encrypt(value, p.aad.Get(tableEncryption, columnValue, name), p.keys.encryption); err != nil {
		return err
	}

	switch {
	case replace:
		_, err = conn.ExecContext(ctx, p.sqlUpsertEncryptionValue, name, value)
	default:
		_, err = conn.ExecContext(ctx, p.sqlInsertEncryptionValue, name, value)
	}

	return err
}

func (p *SQLProvider) setNewEncryptionCheckValue(ctx context.Context, conn SQLXConnection, key []byte, aad EncryptionAAD) (err error) {
	valueClearText, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	value, err := utils.Encrypt([]byte(valueClearText.String()), aad.Get(tableEncryption, columnValue, encryptionNameCheck), key)
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, p.sqlUpsertEncryptionValue, encryptionNameCheck, value)

	return err
}
