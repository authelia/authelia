// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestSessionIDCollectorShouldDeleteMappingsWithNoSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	record := model.OAuth2SessionID{Issuer: "https://example.com", PublicID: "pid1", SessionID: uuid.New()}

	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 0, 100).Return([]model.OAuth2SessionID{record}, nil)
	storage.EXPECT().DeleteOAuth2SessionID(gomock.Any(), record.Issuer, record.SessionID.String()).Return(nil)

	lookup := &mockSessionIDLookup{found: false}

	collector := oidc.NewSessionIDCollector(storage, lookup)

	deleted, err := collector.Collect(context.Background(), 100, 1000)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

func TestSessionIDCollectorShouldKeepMappingsWithLiveSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	record := model.OAuth2SessionID{Issuer: "https://example.com", PublicID: "pid1", SessionID: uuid.New()}

	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 0, 100).Return([]model.OAuth2SessionID{record}, nil)

	lookup := &mockSessionIDLookup{found: true}

	collector := oidc.NewSessionIDCollector(storage, lookup)

	deleted, err := collector.Collect(context.Background(), 100, 1000)

	require.NoError(t, err)
	assert.Equal(t, 0, deleted)
}

func TestSessionIDCollectorShouldDeleteNothingOnRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	records := []model.OAuth2SessionID{
		{Issuer: "https://example.com", PublicID: "pid1", SessionID: uuid.New()},
		{Issuer: "https://example.com", PublicID: "pid2", SessionID: uuid.New()},
	}

	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 0, 100).Return(records, nil)

	lookup := &mockSessionIDLookup{err: errors.New("connection refused")}

	collector := oidc.NewSessionIDCollector(storage, lookup)

	deleted, err := collector.Collect(context.Background(), 100, 1000)

	assert.Error(t, err)
	assert.Equal(t, 0, deleted)
}

func TestSessionIDCollectorShouldPropagateLoadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 0, 100).Return(nil, errors.New("connection refused"))

	lookup := &mockSessionIDLookup{}

	collector := oidc.NewSessionIDCollector(storage, lookup)

	deleted, err := collector.Collect(context.Background(), 100, 1000)

	assert.Error(t, err)
	assert.Equal(t, 0, deleted)
}

func TestSessionIDCollectorShouldAdvanceCursorPastFirstPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	live1 := model.OAuth2SessionID{ID: 1, Issuer: "https://example.com", PublicID: "pid1", SessionID: uuid.New()}
	live2 := model.OAuth2SessionID{ID: 2, Issuer: "https://example.com", PublicID: "pid2", SessionID: uuid.New()}
	orphan := model.OAuth2SessionID{ID: 3, Issuer: "https://example.com", PublicID: "pid3", SessionID: uuid.New()}

	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 0, 2).Return([]model.OAuth2SessionID{live1, live2}, nil)
	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 2, 2).Return([]model.OAuth2SessionID{orphan}, nil)
	storage.EXPECT().DeleteOAuth2SessionID(gomock.Any(), orphan.Issuer, orphan.SessionID.String()).Return(nil)

	lookup := &mockSessionIDLookupTable{found: map[string]bool{live1.PublicID: true, live2.PublicID: true, orphan.PublicID: false}}

	collector := oidc.NewSessionIDCollector(storage, lookup)

	deleted, err := collector.Collect(context.Background(), 2, 1000)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted, "the orphan beyond the first page must be reached and deleted")
}

func TestSessionIDCollectorShouldReturnDeletesCompletedAndErrorWhenLookupFailsAfterSomeDeletes(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	orphan := model.OAuth2SessionID{ID: 1, Issuer: "https://example.com", PublicID: "pid1", SessionID: uuid.New()}
	unresolvable := model.OAuth2SessionID{ID: 2, Issuer: "https://example.com", PublicID: "pid2", SessionID: uuid.New()}

	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 0, 100).Return([]model.OAuth2SessionID{orphan, unresolvable}, nil)
	storage.EXPECT().DeleteOAuth2SessionID(gomock.Any(), orphan.Issuer, orphan.SessionID.String()).Return(nil)

	expected := errors.New("connection refused")

	lookup := &mockSessionIDLookupTable{
		found: map[string]bool{orphan.PublicID: false},
		err:   map[string]error{unresolvable.PublicID: expected},
	}

	collector := oidc.NewSessionIDCollector(storage, lookup)

	deleted, err := collector.Collect(context.Background(), 100, 1000)

	assert.ErrorIs(t, err, expected)
	assert.Greater(t, deleted, 0)
	assert.Equal(t, 1, deleted)
}

func TestSessionIDCollectorShouldResumeAfterHittingRowCap(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	live1 := model.OAuth2SessionID{ID: 1, Issuer: "https://example.com", PublicID: "pid1", SessionID: uuid.New()}
	live2 := model.OAuth2SessionID{ID: 2, Issuer: "https://example.com", PublicID: "pid2", SessionID: uuid.New()}
	orphan := model.OAuth2SessionID{ID: 3, Issuer: "https://example.com", PublicID: "pid3", SessionID: uuid.New()}
	live4 := model.OAuth2SessionID{ID: 4, Issuer: "https://example.com", PublicID: "pid4", SessionID: uuid.New()}

	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 0, 2).Return([]model.OAuth2SessionID{live1, live2}, nil)
	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 2, 2).Return([]model.OAuth2SessionID{orphan, live4}, nil)
	storage.EXPECT().DeleteOAuth2SessionID(gomock.Any(), orphan.Issuer, orphan.SessionID.String()).Return(nil)

	lookup := &mockSessionIDLookupTable{found: map[string]bool{
		live1.PublicID:  true,
		live2.PublicID:  true,
		orphan.PublicID: false,
		live4.PublicID:  true,
	}}

	collector := oidc.NewSessionIDCollector(storage, lookup)

	deleted, err := collector.Collect(context.Background(), 2, 2)

	require.NoError(t, err)
	assert.Equal(t, 0, deleted, "the row cap is hit after the first page, before the orphan on the second page is reached")

	deleted, err = collector.Collect(context.Background(), 2, 2)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted, "the second call must resume at id 2 rather than restart at 0")
}

func TestSessionIDCollectorShouldResumeAfterErrorRatherThanRestart(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	live1 := model.OAuth2SessionID{ID: 1, Issuer: "https://example.com", PublicID: "pid1", SessionID: uuid.New()}
	live2 := model.OAuth2SessionID{ID: 2, Issuer: "https://example.com", PublicID: "pid2", SessionID: uuid.New()}
	orphan := model.OAuth2SessionID{ID: 3, Issuer: "https://example.com", PublicID: "pid3", SessionID: uuid.New()}

	// The after=2 page is fetched twice: once when the backend is flaky and fails the lookup, and once when it has
	// recovered. The after=0 page must only ever be fetched once, proving the second call resumed rather than
	// re-scanning ids 1-2 from the start.
	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 0, 2).Return([]model.OAuth2SessionID{live1, live2}, nil)
	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 2, 2).Return([]model.OAuth2SessionID{orphan}, nil).Times(2)
	storage.EXPECT().DeleteOAuth2SessionID(gomock.Any(), orphan.Issuer, orphan.SessionID.String()).Return(nil)

	expected := errors.New("connection refused")

	lookup := &mockSessionIDLookupTable{
		found: map[string]bool{live1.PublicID: true, live2.PublicID: true},
		err:   map[string]error{orphan.PublicID: expected},
	}

	collector := oidc.NewSessionIDCollector(storage, lookup)

	deleted, err := collector.Collect(context.Background(), 2, 1000)

	assert.ErrorIs(t, err, expected)
	assert.Equal(t, 0, deleted)

	// The backend recovers.
	delete(lookup.err, orphan.PublicID)
	lookup.found[orphan.PublicID] = false

	deleted, err = collector.Collect(context.Background(), 2, 1000)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted, "the second call must resume at id 2 rather than re-scanning ids 1-2 from the start")
}

func TestSessionIDCollectorShouldResetCursorAfterReachingEnd(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	record := model.OAuth2SessionID{ID: 1, Issuer: "https://example.com", PublicID: "pid1", SessionID: uuid.New()}

	// after=0 is expected twice: a sweep which reaches the end of the table on the first call must reset so the
	// second call starts over from the beginning, otherwise a collector which ever finished a pass would never
	// collect a mapping created afterwards.
	storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 0, 100).Return([]model.OAuth2SessionID{record}, nil).Times(2)

	lookup := &mockSessionIDLookup{found: true}

	collector := oidc.NewSessionIDCollector(storage, lookup)

	deleted, err := collector.Collect(context.Background(), 100, 1000)

	require.NoError(t, err)
	assert.Equal(t, 0, deleted)

	deleted, err = collector.Collect(context.Background(), 100, 1000)

	require.NoError(t, err)
	assert.Equal(t, 0, deleted, "a second call after the table was fully swept must start over from 0")
}

type mockSessionIDLookup struct {
	found bool
	err   error
}

func (m *mockSessionIDLookup) GetByPublicID(_ context.Context, _, _ string) (found bool, err error) {
	return m.found, m.err
}

// mockSessionIDLookupTable resolves the found/error result per public id, which multi-page and partial-failure tests
// need in place of a single fixed response for every call.
type mockSessionIDLookupTable struct {
	found map[string]bool
	err   map[string]error
}

func (m *mockSessionIDLookupTable) GetByPublicID(_ context.Context, _, pid string) (found bool, err error) {
	if e, ok := m.err[pid]; ok {
		return false, e
	}

	return m.found[pid], nil
}
