// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/model"
)

func (p *SQLProvider) getHMACRecoveryCode(ctx context.Context) (key []byte, err error) {
	return p.getHMACKey(ctx, "rc", sha256.BlockSize)
}

func (p *SQLProvider) recoveryCodeHMACSignature(values ...[]byte) string {
	h := hmac.New(sha256.New, p.keys.recoveryCodeHMAC)

	for _, v := range values {
		h.Write(v)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// SaveRecoveryCode saves a recovery code to the storage provider after computing its HMAC signature
// from the transient plaintext on the model. The plaintext is never persisted.
func (p *SQLProvider) SaveRecoveryCode(ctx context.Context, code *model.RecoveryCode) (err error) {
	if code.Plaintext == "" {
		return fmt.Errorf("error inserting recovery code for user '%s': plaintext is empty", code.Username)
	}

	code.Signature = p.recoveryCodeHMACSignature([]byte(code.Username), []byte(model.NormalizeRecoveryCode(code.Plaintext)))

	if _, err = p.db.ExecContext(ctx, p.sqlInsertRecoveryCode, code.Username, code.Signature, code.CreatedAt); err != nil {
		return fmt.Errorf("error inserting recovery code for user '%s': %w", code.Username, err)
	}

	return nil
}

// LoadRecoveryCode loads a recovery code from the storage provider given a username and a user-supplied
// raw code. The raw code is normalized and HMAC'd before lookup.
func (p *SQLProvider) LoadRecoveryCode(ctx context.Context, username, raw string) (code *model.RecoveryCode, err error) {
	code = &model.RecoveryCode{}

	signature := p.recoveryCodeHMACSignature([]byte(username), []byte(model.NormalizeRecoveryCode(raw)))

	if err = p.db.GetContext(ctx, code, p.sqlSelectRecoveryCodeBySignatureAndUsername, signature, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting recovery code: %w", err)
	}

	return code, nil
}

// LoadRecoveryCodesByUsername loads all recovery codes for a given username, including consumed and revoked rows.
func (p *SQLProvider) LoadRecoveryCodesByUsername(ctx context.Context, username string) (codes []model.RecoveryCode, err error) {
	codes = nil

	if err = p.db.SelectContext(ctx, &codes, p.sqlSelectRecoveryCodesByUsername, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting recovery codes for user '%s': %w", username, err)
	}

	return codes, nil
}

// ConsumeRecoveryCode marks a recovery code as consumed by stamping the consumption time and remote IP.
func (p *SQLProvider) ConsumeRecoveryCode(ctx context.Context, id int, ip model.NullIP) (err error) {
	consumedAt := sql.NullTime{Valid: true, Time: time.Now()}

	if _, err = p.db.ExecContext(ctx, p.sqlConsumeRecoveryCode, consumedAt, ip, id); err != nil {
		return fmt.Errorf("error updating recovery code (consume) id '%d': %w", id, err)
	}

	return nil
}

// RevokeRecoveryCodesByUsername bulk-soft-deletes all unconsumed recovery codes for a username.
func (p *SQLProvider) RevokeRecoveryCodesByUsername(ctx context.Context, username string, ip model.NullIP) (err error) {
	revokedAt := sql.NullTime{Valid: true, Time: time.Now()}

	if _, err = p.db.ExecContext(ctx, p.sqlRevokeRecoveryCodesByUsername, revokedAt, ip, username); err != nil {
		return fmt.Errorf("error revoking recovery codes for user '%s': %w", username, err)
	}

	return nil
}

// CountUnusedRecoveryCodesByUsername returns the count of recovery codes for a user that are neither consumed nor revoked.
func (p *SQLProvider) CountUnusedRecoveryCodesByUsername(ctx context.Context, username string) (count int, err error) {
	if err = p.db.GetContext(ctx, &count, p.sqlCountUnusedRecoveryCodesByUsername, username); err != nil {
		return 0, fmt.Errorf("error counting unused recovery codes for user '%s': %w", username, err)
	}

	return count, nil
}

// CountUsersWithRecoveryCodes returns the count of distinct users that have at least one unused recovery code.
func (p *SQLProvider) CountUsersWithRecoveryCodes(ctx context.Context) (count int, err error) {
	if err = p.db.GetContext(ctx, &count, p.sqlCountUsersWithRecoveryCodes); err != nil {
		return 0, fmt.Errorf("error counting users with recovery codes: %w", err)
	}

	return count, nil
}

// CountUsersWithLowRecoveryCodes returns the count of distinct users that have between 1 and 2 unused recovery codes.
func (p *SQLProvider) CountUsersWithLowRecoveryCodes(ctx context.Context) (count int, err error) {
	if err = p.db.GetContext(ctx, &count, p.sqlCountUsersWithLowRecoveryCodes); err != nil {
		return 0, fmt.Errorf("error counting users with low recovery codes: %w", err)
	}

	return count, nil
}

// CountUsersWithDepletedRecoveryCodes returns the count of distinct users that have zero unused recovery codes
// but at least one consumed or revoked code.
func (p *SQLProvider) CountUsersWithDepletedRecoveryCodes(ctx context.Context) (count int, err error) {
	if err = p.db.GetContext(ctx, &count, p.sqlCountUsersWithDepletedRecoveryCodes); err != nil {
		return 0, fmt.Errorf("error counting users with depleted recovery codes: %w", err)
	}

	return count, nil
}
