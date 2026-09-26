// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/cache"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

var errTestSessionBackend = errors.New("backend unavailable")

func newTestSessionCookie(label string) string {
	id := sha256.Sum256([]byte(label))

	return base64.RawURLEncoding.EncodeToString(id[:])
}

type failingSessionRepository struct {
	session.Repository

	errGet      error
	errDelete   error
	errChangeID error

	errSave []error
	saved   []string
}

func (r *failingSessionRepository) Get(ctx context.Context, issuer, id string) (record session.Record, err error) {
	if r.errGet != nil {
		return nil, r.errGet
	}

	return r.Repository.Get(ctx, issuer, id)
}

func (r *failingSessionRepository) Save(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	if len(r.errSave) != 0 {
		err, r.errSave = r.errSave[0], r.errSave[1:]
		if err != nil {
			return err
		}
	}

	if err = r.Repository.Save(ctx, issuer, id, pid, username, expiration, data); err != nil {
		return err
	}

	r.saved = append(r.saved, username)

	return nil
}

func (r *failingSessionRepository) Delete(ctx context.Context, issuer, id, pid, username string) (err error) {
	if r.errDelete != nil {
		return r.errDelete
	}

	return r.Repository.Delete(ctx, issuer, id, pid, username)
}

func (r *failingSessionRepository) ChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	if r.errChangeID != nil {
		return r.errChangeID
	}

	return r.Repository.ChangeID(ctx, issuer, oldID, id, pid, username, expiration, data)
}

func setupTestFailingSessionRepository(t *testing.T, mock *mocks.MockAutheliaCtx) (repository *failingSessionRepository) {
	t.Helper()

	repository = &failingSessionRepository{Repository: cache.NewSessionRepository(cache.NewMemory())}

	provider, err := session.NewProvider(&mock.Ctx.Configuration, []byte("cd21a70d4d24b23e0d1ae0a5cf6c4e1d3e2b8a5f7c9d0e1f2a3b4c5d6e7f8a9b"), mock.Ctx.Providers.Clock, mock.Ctx.Providers.Random, repository)
	require.NoError(t, err)

	mock.Ctx.Providers.Session = provider

	return repository
}
