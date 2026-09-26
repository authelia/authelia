// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageSessionRepositoryShouldDelegateToTheProvider(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	repository := NewSessionRepository(provider)

	const (
		issuer    = "an-issuer"
		signature = "a-signature"
		moved     = "a-moved-signature"
		publicID  = "a-public-id"
		username  = "john"
	)

	require.NoError(t, repository.StartupCheck())

	require.NoError(t, repository.Save(ctx, issuer, signature, publicID, username, time.Hour, []byte("first")))

	record, err := repository.Get(ctx, issuer, signature)

	require.NoError(t, err)
	require.NotNil(t, record)
	assert.Equal(t, []byte("first"), record.GetSessionData())

	require.NoError(t, repository.SaveData(ctx, issuer, signature, publicID, username, time.Hour, []byte("second")))

	record, err = repository.GetByPublicID(ctx, issuer, publicID)

	require.NoError(t, err)
	require.NotNil(t, record)
	assert.Equal(t, []byte("second"), record.GetSessionData())

	ids, err := repository.GetIDsByUsername(ctx, issuer, username)

	require.NoError(t, err)
	assert.Equal(t, []string{signature}, ids)

	require.NoError(t, repository.ChangeID(ctx, issuer, signature, moved, publicID, username, time.Hour, []byte("third")))

	record, err = repository.Get(ctx, issuer, moved)

	require.NoError(t, err)
	require.NotNil(t, record)
	assert.Equal(t, []byte("third"), record.GetSessionData())

	require.NoError(t, repository.GarbageCollection(ctx))
	assert.Equal(t, sessionGarbageCollectionFrequency, repository.GarbageCollectionFrequency(ctx))

	require.NoError(t, repository.Delete(ctx, issuer, moved, publicID, username))

	record, err = repository.Get(ctx, issuer, moved)

	require.NoError(t, err)
	assert.Nil(t, record)
}

func TestStorageSessionShouldReturnErrorsWhenTheDatabaseIsUnavailable(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.Close())

	record, err := provider.SessionGet(ctx, "an-issuer", "a-signature")

	assert.Nil(t, record)
	assert.ErrorContains(t, err, "error selecting session: ")

	record, err = provider.SessionGetByPublicID(ctx, "an-issuer", "a-public-id")

	assert.Nil(t, record)
	assert.ErrorContains(t, err, "error selecting session by public id: ")

	ids, err := provider.SessionGetIDsByUsername(ctx, "an-issuer", "john")

	assert.Nil(t, ids)
	assert.ErrorContains(t, err, "error selecting session ids by username: ")

	assert.ErrorContains(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("data")), "error upserting session: ")
	assert.ErrorContains(t, provider.SessionSaveData(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("data")), "error updating session data: ")
	assert.ErrorContains(t, provider.SessionDelete(ctx, "an-issuer", "a-signature", "a-public-id", "john"), "error deleting session: ")
	assert.ErrorContains(t, provider.SessionChangeID(ctx, "an-issuer", "a-signature", "another-signature", "a-public-id", "john", time.Hour, []byte("data")), "error updating session signature: ")
	assert.ErrorContains(t, provider.SessionGarbageCollection(ctx), "error deleting expired sessions: ")
}

func TestStorageSessionSaveShouldReturnTheErrorWhenThePublicIDLookupFails(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("data")))

	provider.sqlSelectSessionSignatureByPublicID = "SELECT signature FROM a_table_which_does_not_exist WHERE issuer = ? AND public_id = ?;"

	err := provider.SessionSave(ctx, "an-issuer", "another-signature", "a-public-id", "john", time.Hour, []byte("data"))

	assert.ErrorContains(t, err, "error upserting session: ")
	assert.ErrorContains(t, err, "UNIQUE constraint failed")
}

func TestSQLProviderLoadHMACKeyShouldPersistTheKey(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	key, err := provider.LoadHMACKey(ctx, "a-name", 64)

	require.NoError(t, err)
	assert.Len(t, key, 64)

	again, err := provider.LoadHMACKey(ctx, "a-name", 64)

	require.NoError(t, err)
	assert.Equal(t, key, again)
}

func TestSQLProviderLoadHMACKeyShouldAdoptTheKeyCommittedByAnotherInstance(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	expected, err := provider.LoadHMACKey(ctx, "a-name", 64)
	require.NoError(t, err)

	key, err := provider.getHMACKeyExisting(ctx, "a-name")

	require.NoError(t, err)
	assert.Equal(t, expected, key)
}

func TestSQLProviderLoadHMACKeyShouldErrorWhenTheCommittedKeyIsAbsent(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	key, err := provider.getHMACKeyExisting(ctx, "a-name-which-was-never-committed")

	assert.Error(t, err)
	assert.Nil(t, key)
}
