package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/model"
)

// SaveTelegramVerification saves a pending Telegram verification to the storage provider.
func (p *SQLProvider) SaveTelegramVerification(ctx context.Context, verification model.TelegramVerification) (err error) {
	if verification.CreatedAt.IsZero() {
		verification.CreatedAt = time.Now()
	}

	if _, err = p.db.ExecContext(ctx, p.sqlInsertTelegramVerification,
		verification.Username, verification.Token, verification.TelegramID,
		verification.Phone, verification.Verified, verification.CreatedAt); err != nil {
		return fmt.Errorf("error inserting telegram verification for user '%s': %w", verification.Username, err)
	}

	return nil
}

// LoadTelegramVerification loads a Telegram verification with TTL enforcement.
func (p *SQLProvider) LoadTelegramVerification(ctx context.Context, username, token string, createdAfter time.Time) (verification *model.TelegramVerification, err error) {
	verification = &model.TelegramVerification{}

	if err = p.db.QueryRowxContext(ctx, p.sqlSelectTelegramVerification, username, token, createdAfter).StructScan(verification); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoTelegramVerification
		}

		return nil, fmt.Errorf("error selecting telegram verification for user '%s': %w", username, err)
	}

	return verification, nil
}

// DeleteTelegramVerification deletes a Telegram verification for a given username and token.
func (p *SQLProvider) DeleteTelegramVerification(ctx context.Context, username, token string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteTelegramVerification, username, token); err != nil {
		return fmt.Errorf("error deleting telegram verification for user '%s': %w", username, err)
	}

	return nil
}

// DeleteTelegramVerificationsPending deletes all pending (unverified) tokens for a user.
func (p *SQLProvider) DeleteTelegramVerificationsPending(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteTelegramVerificationsPending, username); err != nil {
		return fmt.Errorf("error deleting pending telegram verifications for user '%s': %w", username, err)
	}

	return nil
}

// DeleteTelegramVerificationsExpired deletes all expired tokens globally.
func (p *SQLProvider) DeleteTelegramVerificationsExpired(ctx context.Context, before time.Time) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteTelegramVerificationsExpired, before); err != nil {
		return fmt.Errorf("error deleting expired telegram verifications: %w", err)
	}

	return nil
}
