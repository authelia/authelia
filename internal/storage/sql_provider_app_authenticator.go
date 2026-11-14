// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/model"
)

const appAuthenticatorRetention = time.Hour * 24

func utcNullTime(t sql.NullTime) sql.NullTime {
	if t.Valid {
		t.Time = t.Time.UTC()
	}

	return t
}

// SaveAppAuthenticatorDevice saves a new Authelia app device to the storage provider.
func (p *SQLProvider) SaveAppAuthenticatorDevice(ctx context.Context, d model.AppAuthenticatorDevice) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlInsertAppAuthenticatorDevice,
		d.CreatedAt.UTC(), utcNullTime(d.LastUsedAt), d.DeviceID, d.Username, d.BaseURL, d.Name, d.Model, d.Platform,
		d.ApprovalPublicKey, d.TransportPublicKey, d.AttestationFormat, d.AttestationIdentity, d.AttestationDetail,
		d.AttestationKeyID, d.PullIDCurrent, d.PullIDPrevious,
	); err != nil {
		return fmt.Errorf("error inserting app authenticator device for user '%s': %w", d.Username, err)
	}

	return nil
}

// LoadAppAuthenticatorDeviceByDeviceID loads an Authelia app device by its public device id.
func (p *SQLProvider) LoadAppAuthenticatorDeviceByDeviceID(ctx context.Context, deviceID string) (device *model.AppAuthenticatorDevice, err error) {
	device = &model.AppAuthenticatorDevice{}

	if err = p.conn(ctx).QueryRowxContext(ctx, p.sqlSelectAppAuthenticatorDeviceByDeviceID, deviceID).StructScan(device); err != nil {
		return nil, fmt.Errorf("error selecting app authenticator device: %w", err)
	}

	return device, nil
}

// LoadAppAuthenticatorDeviceByID loads an Authelia app device by its row id.
func (p *SQLProvider) LoadAppAuthenticatorDeviceByID(ctx context.Context, id int) (device *model.AppAuthenticatorDevice, err error) {
	device = &model.AppAuthenticatorDevice{}

	if err = p.conn(ctx).QueryRowxContext(ctx, p.sqlSelectAppAuthenticatorDeviceByID, id).StructScan(device); err != nil {
		return nil, fmt.Errorf("error selecting app authenticator device with id '%d': %w", id, err)
	}

	return device, nil
}

// LoadAppAuthenticatorDevicesByUsername loads all Authelia app devices for a user.
func (p *SQLProvider) LoadAppAuthenticatorDevicesByUsername(ctx context.Context, username string) (devices []model.AppAuthenticatorDevice, err error) {
	if err = p.conn(ctx).SelectContext(ctx, &devices, p.sqlSelectAppAuthenticatorDevicesByUsername, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting app authenticator devices for user '%s': %w", username, err)
	}

	return devices, nil
}

// UpdateAppAuthenticatorDevicePullIDs rotates the pull id hashes of a device if the current hash still equals expectedCurrent.
func (p *SQLProvider) UpdateAppAuthenticatorDevicePullIDs(ctx context.Context, id int, expectedCurrent, current, previous []byte) (updated bool, err error) {
	var result sql.Result

	if result, err = p.conn(ctx).ExecContext(ctx, p.sqlUpdateAppAuthenticatorDevicePullIDs, current, previous, id, expectedCurrent); err != nil {
		return false, fmt.Errorf("error updating app authenticator device pull ids for id '%d': %w", id, err)
	}

	var affected int64

	if affected, err = result.RowsAffected(); err != nil {
		return false, fmt.Errorf("error updating app authenticator device pull ids for id '%d': %w", id, err)
	}

	return affected > 0, nil
}

// UpdateAppAuthenticatorDeviceName updates the display name of an Authelia app device owned by the user.
func (p *SQLProvider) UpdateAppAuthenticatorDeviceName(ctx context.Context, username string, id int, name string) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlUpdateAppAuthenticatorDeviceName, name, username, id); err != nil {
		return fmt.Errorf("error updating app authenticator device name for id '%d': %w", id, err)
	}

	return nil
}

// UpdateAppAuthenticatorDeviceLastUsed updates the last used time of an Authelia app device.
func (p *SQLProvider) UpdateAppAuthenticatorDeviceLastUsed(ctx context.Context, id int, at time.Time) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlUpdateAppAuthenticatorDeviceLastUsed, at.UTC(), id); err != nil {
		return fmt.Errorf("error updating app authenticator device last used for id '%d': %w", id, err)
	}

	return nil
}

// DeleteAppAuthenticatorDevice deletes an Authelia app device, its request nonces and clears pending challenge bindings to it.
func (p *SQLProvider) DeleteAppAuthenticatorDevice(ctx context.Context, deviceID string) (err error) {
	if _, ok := ctx.Value(ctxKeyTransaction).(SQLXTx); !ok {
		if ctx, err = p.BeginTX(ctx); err != nil {
			return fmt.Errorf("error beginning transaction to delete app authenticator device: %w", err)
		}

		defer func() {
			if err != nil {
				_ = p.Rollback(ctx)

				return
			}

			if err = p.Commit(ctx); err != nil {
				err = fmt.Errorf("error committing transaction to delete app authenticator device: %w", err)
			}
		}()
	}

	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlDeleteAppAuthenticatorRequestNoncesByDeviceID, deviceID); err != nil {
		return fmt.Errorf("error deleting app authenticator request nonces: %w", err)
	}

	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlClearAppAuthenticatorChallengeBindings, deviceID); err != nil {
		return fmt.Errorf("error clearing app authenticator challenge bindings: %w", err)
	}

	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlDeleteAppAuthenticatorDevice, deviceID); err != nil {
		return fmt.Errorf("error deleting app authenticator device: %w", err)
	}

	return nil
}

// SaveAppAuthenticatorRegistration saves a new Authelia app registration to the storage provider.
func (p *SQLProvider) SaveAppAuthenticatorRegistration(ctx context.Context, r model.AppAuthenticatorRegistration) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlInsertAppAuthenticatorRegistration,
		r.CreatedAt.UTC(), r.ExpiresAt.UTC(), r.Code, r.Username, r.PortalBinding, r.BaseURL, r.Number, r.NonceLength, r.State,
		utcNullTime(r.BegunAt), utcNullTime(r.SessionExpiresAt), r.ConfirmationID, r.Challenge, r.CodeChallenge,
		r.DevicePlatform, r.DeviceModel, r.DeviceName, r.DeviceID,
	); err != nil {
		return fmt.Errorf("error inserting app authenticator registration for user '%s': %w", r.Username, err)
	}

	return nil
}

// LoadAppAuthenticatorRegistrationByCode loads an Authelia app registration by the hash of its code.
func (p *SQLProvider) LoadAppAuthenticatorRegistrationByCode(ctx context.Context, code []byte) (registration *model.AppAuthenticatorRegistration, err error) {
	return p.loadAppAuthenticatorRegistration(ctx, p.sqlSelectAppAuthenticatorRegistrationByCode, code)
}

// LoadAppAuthenticatorRegistrationByConfirmationID loads an Authelia app registration by the hash of its confirmation id.
func (p *SQLProvider) LoadAppAuthenticatorRegistrationByConfirmationID(ctx context.Context, id []byte) (registration *model.AppAuthenticatorRegistration, err error) {
	return p.loadAppAuthenticatorRegistration(ctx, p.sqlSelectAppAuthenticatorRegistrationByConfirmationID, id)
}

// LoadAppAuthenticatorRegistrationByPortalBinding loads the newest Authelia app registration for a portal binding.
func (p *SQLProvider) LoadAppAuthenticatorRegistrationByPortalBinding(ctx context.Context, binding []byte) (registration *model.AppAuthenticatorRegistration, err error) {
	return p.loadAppAuthenticatorRegistration(ctx, p.sqlSelectAppAuthenticatorRegistrationByPortalBinding, binding)
}

func (p *SQLProvider) loadAppAuthenticatorRegistration(ctx context.Context, query string, arg []byte) (registration *model.AppAuthenticatorRegistration, err error) {
	registration = &model.AppAuthenticatorRegistration{}

	if err = p.conn(ctx).QueryRowxContext(ctx, query, arg).StructScan(registration); err != nil {
		return nil, fmt.Errorf("error selecting app authenticator registration: %w", err)
	}

	return registration, nil
}

// UpdateAppAuthenticatorRegistration updates the mutable columns of an Authelia app registration by id only if it is still in the expected state.
func (p *SQLProvider) UpdateAppAuthenticatorRegistration(ctx context.Context, r model.AppAuthenticatorRegistration, expectedState string) (updated bool, err error) {
	var result sql.Result

	if result, err = p.conn(ctx).ExecContext(ctx, p.sqlUpdateAppAuthenticatorRegistration,
		r.ExpiresAt.UTC(), r.PortalBinding, r.State, utcNullTime(r.BegunAt), utcNullTime(r.SessionExpiresAt), r.ConfirmationID,
		r.Challenge, r.CodeChallenge, r.DevicePlatform, r.DeviceModel, r.DeviceName, r.DeviceID, r.ID, expectedState,
	); err != nil {
		return false, fmt.Errorf("error updating app authenticator registration with id '%d': %w", r.ID, err)
	}

	return rowsAffectedPositive(result, "app authenticator registration")
}

func rowsAffectedPositive(result sql.Result, what string) (ok bool, err error) {
	var affected int64

	if affected, err = result.RowsAffected(); err != nil {
		return false, fmt.Errorf("error updating %s: %w", what, err)
	}

	return affected > 0, nil
}

// CancelAppAuthenticatorRegistrations cancels all non-terminal Authelia app registrations for a portal binding.
func (p *SQLProvider) CancelAppAuthenticatorRegistrations(ctx context.Context, binding []byte) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlCancelAppAuthenticatorRegistrations, binding); err != nil {
		return fmt.Errorf("error cancelling app authenticator registrations: %w", err)
	}

	return nil
}

// SaveAppAuthenticatorChallenge saves a new Authelia app challenge to the storage provider.
func (p *SQLProvider) SaveAppAuthenticatorChallenge(ctx context.Context, c model.AppAuthenticatorChallenge) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlInsertAppAuthenticatorChallenge,
		c.CreatedAt.UTC(), c.ExpiresAt.UTC(), c.ConfirmationID, c.Username, c.PortalBinding, c.Challenge, c.Number, c.NonceLength, c.State,
		c.BoundDeviceID, c.BoundCodeChallenge, utcNullTime(c.DecidedAt), c.DecidedDeviceID, c.PortalURL, c.RequestedURL,
		c.RemoteIP, c.UserAgent, utcNullTime(c.ConsumedAt),
	); err != nil {
		return fmt.Errorf("error inserting app authenticator challenge for user '%s': %w", c.Username, err)
	}

	return nil
}

// LoadAppAuthenticatorChallengeByConfirmationID loads an Authelia app challenge by its confirmation id.
func (p *SQLProvider) LoadAppAuthenticatorChallengeByConfirmationID(ctx context.Context, id string) (challenge *model.AppAuthenticatorChallenge, err error) {
	challenge = &model.AppAuthenticatorChallenge{}

	if err = p.conn(ctx).QueryRowxContext(ctx, p.sqlSelectAppAuthenticatorChallengeByConfirmationID, id).StructScan(challenge); err != nil {
		return nil, fmt.Errorf("error selecting app authenticator challenge: %w", err)
	}

	return challenge, nil
}

// LoadAppAuthenticatorChallengeByPortalBinding loads the newest Authelia app challenge for a portal binding.
func (p *SQLProvider) LoadAppAuthenticatorChallengeByPortalBinding(ctx context.Context, binding []byte) (challenge *model.AppAuthenticatorChallenge, err error) {
	challenge = &model.AppAuthenticatorChallenge{}

	if err = p.conn(ctx).QueryRowxContext(ctx, p.sqlSelectAppAuthenticatorChallengeByPortalBinding, binding).StructScan(challenge); err != nil {
		return nil, fmt.Errorf("error selecting app authenticator challenge: %w", err)
	}

	return challenge, nil
}

// LoadAppAuthenticatorChallengePending loads the oldest pending unexpired Authelia app challenge for a user.
func (p *SQLProvider) LoadAppAuthenticatorChallengePending(ctx context.Context, username string, now time.Time) (challenge *model.AppAuthenticatorChallenge, err error) {
	challenge = &model.AppAuthenticatorChallenge{}

	if err = p.conn(ctx).QueryRowxContext(ctx, p.sqlSelectAppAuthenticatorChallengePending, username, now.UTC()).StructScan(challenge); err != nil {
		return nil, fmt.Errorf("error selecting pending app authenticator challenge for user '%s': %w", username, err)
	}

	return challenge, nil
}

// UpdateAppAuthenticatorChallenge updates the mutable columns of an Authelia app challenge by id only if it is still in the expected state.
func (p *SQLProvider) UpdateAppAuthenticatorChallenge(ctx context.Context, c model.AppAuthenticatorChallenge, expectedState string) (updated bool, err error) {
	var result sql.Result

	if result, err = p.conn(ctx).ExecContext(ctx, p.sqlUpdateAppAuthenticatorChallenge,
		c.ExpiresAt.UTC(), c.PortalBinding, c.State, c.BoundDeviceID, c.BoundCodeChallenge, utcNullTime(c.DecidedAt),
		c.DecidedDeviceID, c.PortalURL, c.RequestedURL, c.RemoteIP, c.UserAgent, utcNullTime(c.ConsumedAt), c.ID, expectedState,
	); err != nil {
		return false, fmt.Errorf("error updating app authenticator challenge with id '%d': %w", c.ID, err)
	}

	return rowsAffectedPositive(result, "app authenticator challenge")
}

// UpdateAppAuthenticatorChallengeDecision settles a pending Authelia app challenge with the state, decision time and
// deciding device of the given challenge only if its binding still equals the binding of the given challenge.
func (p *SQLProvider) UpdateAppAuthenticatorChallengeDecision(ctx context.Context, c model.AppAuthenticatorChallenge) (updated bool, err error) {
	var result sql.Result

	if result, err = p.conn(ctx).ExecContext(ctx, p.sqlUpdateAppAuthenticatorChallengeDecision,
		c.State, utcNullTime(c.DecidedAt), c.DecidedDeviceID, c.ID, c.BoundDeviceID.String, c.BoundCodeChallenge.String,
	); err != nil {
		return false, fmt.Errorf("error updating app authenticator challenge decision with id '%d': %w", c.ID, err)
	}

	return rowsAffectedPositive(result, "app authenticator challenge")
}

// ConsumeAppAuthenticatorChallenge marks an approved Authelia app challenge as consumed at the given time, returning
// false if it was not approved, was already consumed, or its expiration is before expiresNotBefore.
func (p *SQLProvider) ConsumeAppAuthenticatorChallenge(ctx context.Context, id int, at, expiresNotBefore time.Time) (consumed bool, err error) {
	var result sql.Result

	if result, err = p.conn(ctx).ExecContext(ctx, p.sqlConsumeAppAuthenticatorChallenge, at.UTC(), id, expiresNotBefore.UTC()); err != nil {
		return false, fmt.Errorf("error consuming app authenticator challenge with id '%d': %w", id, err)
	}

	return rowsAffectedPositive(result, "app authenticator challenge")
}

// VoidAppAuthenticatorChallengesByPortalBinding voids all pending Authelia app challenges for a portal binding.
func (p *SQLProvider) VoidAppAuthenticatorChallengesByPortalBinding(ctx context.Context, binding []byte) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlVoidAppAuthenticatorChallengesByPortalBinding, binding); err != nil {
		return fmt.Errorf("error voiding app authenticator challenges: %w", err)
	}

	return nil
}

// SaveAppAuthenticatorRequestNonce records a request nonce, returning false if it was already present.
func (p *SQLProvider) SaveAppAuthenticatorRequestNonce(ctx context.Context, deviceID, nonce string, expires time.Time) (inserted bool, err error) {
	var result sql.Result

	if result, err = p.conn(ctx).ExecContext(ctx, p.sqlInsertAppAuthenticatorRequestNonce, deviceID, nonce, expires.UTC()); err != nil {
		return false, fmt.Errorf("error inserting app authenticator request nonce: %w", err)
	}

	var affected int64

	if affected, err = result.RowsAffected(); err != nil {
		return false, fmt.Errorf("error inserting app authenticator request nonce: %w", err)
	}

	return affected > 0, nil
}

// PruneAppAuthenticator deletes expired request nonces and registrations and challenges which expired over a day ago.
func (p *SQLProvider) PruneAppAuthenticator(ctx context.Context, now time.Time) (err error) {
	now = now.UTC()
	cutoff := now.Add(-appAuthenticatorRetention)

	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlDeleteAppAuthenticatorRequestNoncesExpired, now); err != nil {
		return fmt.Errorf("error pruning app authenticator request nonces: %w", err)
	}

	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlDeleteAppAuthenticatorRegistrationsStale, cutoff, cutoff); err != nil {
		return fmt.Errorf("error pruning app authenticator registrations: %w", err)
	}

	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlDeleteAppAuthenticatorChallengesStale, cutoff); err != nil {
		return fmt.Errorf("error pruning app authenticator challenges: %w", err)
	}

	return nil
}
