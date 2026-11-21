// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestSessionIDCollector(t *testing.T) {
	r1 := model.OAuth2SessionID{ID: 1, Issuer: "https://example.com", PublicID: "pid1", SessionID: uuid.New()}
	r2 := model.OAuth2SessionID{ID: 2, Issuer: "https://example.com", PublicID: "pid2", SessionID: uuid.New()}
	r3 := model.OAuth2SessionID{ID: 3, Issuer: "https://example.com", PublicID: "pid3", SessionID: uuid.New()}
	r4 := model.OAuth2SessionID{ID: 4, Issuer: "https://example.com", PublicID: "pid4", SessionID: uuid.New()}
	r1b := model.OAuth2SessionID{ID: 2, Issuer: "https://example.com", SectorID: "b", PublicID: "pid1", SessionID: uuid.New()}

	refused := errors.New("connection refused")

	type run struct {
		ended int
		err   string
	}

	testCases := []struct {
		name      string
		pageSize  int
		maxRows   int
		confirm   time.Duration
		cancelled bool
		found     map[string][]bool
		errs      map[string]error
		setup     func(storage *mocks.MockStorage)
		between   func(lookup *testSessionIDLookup)
		runs      []run
		lookups   int
	}{
		{
			name:  "ShouldEndSessionsWhichNoLongerExist",
			found: map[string][]bool{"pid1": {false}},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 100, r1)
				expectEnd(storage, r1)
			},
			runs: []run{{ended: 1}},
		},
		{
			name:  "ShouldKeepSessionsWhichExist",
			found: map[string][]bool{"pid1": {true}},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 100, r1)
			},
			runs: []run{{}},
		},
		{
			name: "ShouldEndNothingOnRepositoryError",
			errs: map[string]error{"pid1": refused, "pid2": refused},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 100, r1, r2)
			},
			runs: []run{{err: "connection refused"}},
		},
		{
			name: "ShouldPropagateLoadError",
			setup: func(storage *mocks.MockStorage) {
				storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), 0, 100).Return(nil, refused)
			},
			runs: []run{{err: "connection refused"}},
		},
		{
			name:     "ShouldAdvanceCursorPastFirstPage",
			pageSize: 2,
			found:    map[string][]bool{"pid1": {true}, "pid2": {true}, "pid3": {false}},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 2, r1, r2)
				expectSessionIDsOldest(storage, 2, 2, r3)
				expectEnd(storage, r3)
			},
			runs: []run{{ended: 1}},
		},
		{
			name:  "ShouldEndNothingInPageWhenLookupFails",
			found: map[string][]bool{"pid1": {false}},
			errs:  map[string]error{"pid2": refused},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 100, r1, r2)
			},
			runs: []run{{err: "connection refused"}},
		},
		{
			name:     "ShouldResumeAfterHittingRowCap",
			pageSize: 2,
			maxRows:  2,
			found:    map[string][]bool{"pid1": {true}, "pid2": {true}, "pid3": {false}, "pid4": {true}},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 2, r1, r2)
				expectSessionIDsOldest(storage, 2, 2, r3, r4)
				expectEnd(storage, r3)
			},
			runs: []run{{}, {ended: 1}},
		},
		{
			name:     "ShouldResumeAfterErrorRatherThanRestart",
			pageSize: 2,
			found:    map[string][]bool{"pid1": {true}, "pid2": {true}},
			errs:     map[string]error{"pid3": refused},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 2, r1, r2)
				expectSessionIDsOldest(storage, 2, 2, r3).Times(2)
				expectEnd(storage, r3)
			},
			between: func(lookup *testSessionIDLookup) {
				delete(lookup.errs, "pid3")
				lookup.found["pid3"] = []bool{false}
			},
			runs: []run{{err: "connection refused"}, {ended: 1}},
		},
		{
			name:  "ShouldResetCursorAfterReachingEnd",
			found: map[string][]bool{"pid1": {true}},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 100, r1).Times(2)
			},
			runs: []run{{}, {}},
		},
		{
			name:    "ShouldNotEndSessionWhichReappears",
			confirm: time.Millisecond,
			found:   map[string][]bool{"pid1": {false, true}},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 100, r1)
			},
			runs:    []run{{}},
			lookups: 2,
		},
		{
			name:  "ShouldEndEachSessionOnce",
			found: map[string][]bool{"pid1": {false}},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 100, r1, r1b)
				storage.EXPECT().EndOAuth2Session(gomock.Any(), r1.Issuer, r1.PublicID, gomock.Any()).Return([]string{r1.SessionID.String(), r1b.SessionID.String()}, nil)
				storage.EXPECT().RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), r1.SessionID.String()).Return(nil)
				storage.EXPECT().RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), r1b.SessionID.String()).Return(nil)
			},
			runs: []run{{ended: 2}},
		},
		{
			name:  "ShouldContinuePastStorageErrors",
			found: map[string][]bool{"pid1": {false}, "pid2": {false}, "pid3": {false}},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 100, r1, r2, r3)
				storage.EXPECT().EndOAuth2Session(gomock.Any(), r1.Issuer, r1.PublicID, gomock.Any()).Return(nil, errors.New("end failed"))
				storage.EXPECT().EndOAuth2Session(gomock.Any(), r2.Issuer, r2.PublicID, gomock.Any()).Return([]string{r2.SessionID.String()}, nil)
				storage.EXPECT().RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), r2.SessionID.String()).Return(errors.New("revoke failed"))
				expectEnd(storage, r3)
			},
			runs: []run{{ended: 2, err: "end failed\nrevoke failed"}},
		},
		{
			name:      "ShouldStopWaitingWhenCancelled",
			confirm:   time.Hour,
			cancelled: true,
			found:     map[string][]bool{"pid1": {false}},
			setup: func(storage *mocks.MockStorage) {
				expectSessionIDsOldest(storage, 0, 100, r1)
			},
			runs: []run{{err: "context canceled"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockStorage(ctrl)

			if tc.setup != nil {
				tc.setup(storage)
			}

			lookup := &testSessionIDLookup{found: map[string][]bool{}, errs: map[string]error{}, calls: map[string]int{}}

			for pid, found := range tc.found {
				lookup.found[pid] = found
			}

			for pid, err := range tc.errs {
				lookup.errs[pid] = err
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if tc.cancelled {
				cancel()
			}

			pageSize, maxRows := tc.pageSize, tc.maxRows

			if pageSize == 0 {
				pageSize = 100
			}

			if maxRows == 0 {
				maxRows = 1000
			}

			collector := oidc.NewSessionIDCollector(nil, storage, lookup, tc.confirm)

			for i, expected := range tc.runs {
				if i != 0 && tc.between != nil {
					tc.between(lookup)
				}

				ended, err := collector.Collect(ctx, pageSize, maxRows)

				assert.Equal(t, expected.ended, ended)

				if expected.err == "" {
					assert.NoError(t, err)
				} else {
					assert.EqualError(t, err, expected.err)
				}
			}

			if tc.lookups != 0 {
				assert.Equal(t, tc.lookups, lookup.calls["pid1"])
			}
		})
	}
}

func TestSessionIDCollectorFrequency(t *testing.T) {
	collector := oidc.NewSessionIDCollector(nil, nil, nil, 0)

	assert.Equal(t, 5*time.Minute, collector.GarbageCollectionFrequency(context.Background()))
}

func expectSessionIDsOldest(storage *mocks.MockStorage, after, limit int, records ...model.OAuth2SessionID) *gomock.Call {
	return storage.EXPECT().LoadOAuth2SessionIDsOldest(gomock.Any(), after, limit).Return(records, nil)
}

func expectEnd(storage *mocks.MockStorage, record model.OAuth2SessionID) {
	storage.EXPECT().EndOAuth2Session(gomock.Any(), record.Issuer, record.PublicID, gomock.Any()).Return([]string{record.SessionID.String()}, nil)
	storage.EXPECT().RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), record.SessionID.String()).Return(nil)
}

type testSessionIDLookup struct {
	found map[string][]bool
	errs  map[string]error
	calls map[string]int
}

func (m *testSessionIDLookup) GetByPublicID(_ context.Context, _, pid string) (found bool, err error) {
	if err, ok := m.errs[pid]; ok {
		return false, err
	}

	sequence := m.found[pid]
	if len(sequence) == 0 {
		return false, nil
	}

	found = sequence[min(m.calls[pid], len(sequence)-1)]
	m.calls[pid]++

	return found, nil
}
