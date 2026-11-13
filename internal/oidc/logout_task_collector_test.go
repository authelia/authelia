// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestLogoutTaskCollector(t *testing.T) {
	now := time.Unix(1700000000, 0)

	revoke := model.OAuth2SessionLogoutTask{ID: 1, Kind: model.OAuth2SessionLogoutTaskKindRevoke, SessionID: "sid", Attempts: 0, ExpiresAt: now.Add(time.Minute)}
	expired := model.OAuth2SessionLogoutTask{ID: 2, Kind: model.OAuth2SessionLogoutTaskKindRevoke, SessionID: "sid", Attempts: 1, ExpiresAt: now}
	unknown := model.OAuth2SessionLogoutTask{ID: 3, Kind: "unknown", SessionID: "sid", Attempts: 0, ExpiresAt: now}
	retried := model.OAuth2SessionLogoutTask{ID: 4, Kind: "unknown", SessionID: "sid", Attempts: 3, ExpiresAt: now.Add(time.Hour)}
	exhausted := model.OAuth2SessionLogoutTask{ID: 5, Kind: "unknown", SessionID: "sid", Attempts: 100, ExpiresAt: now.Add(time.Hour)}

	testCases := []struct {
		name      string
		tasks     []model.OAuth2SessionLogoutTask
		setup     func(storage *mocks.MockStorage)
		processed int
		err       string
	}{
		{
			name:  "ShouldRescheduleRevokeUntilItExpires",
			tasks: []model.OAuth2SessionLogoutTask{revoke},
			setup: func(storage *mocks.MockStorage) {
				storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 1, 0, now.Add(time.Minute)).Return(true, nil)
				storage.EXPECT().RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), "sid").Return(nil)
				storage.EXPECT().RescheduleOAuth2SessionLogoutTask(gomock.Any(), 1, 1, revoke.ExpiresAt).Return(true, nil)
			},
			processed: 1,
		},
		{
			name:  "ShouldDeleteRevokeOnceExpired",
			tasks: []model.OAuth2SessionLogoutTask{expired},
			setup: func(storage *mocks.MockStorage) {
				storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 2, 1, now.Add(2*time.Minute)).Return(true, nil)
				storage.EXPECT().RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), "sid").Return(nil)
				storage.EXPECT().DeleteOAuth2SessionLogoutTask(gomock.Any(), 2, 2).Return(true, nil)
			},
			processed: 1,
		},
		{
			name:  "ShouldDeleteUnknownKindOnceExpired",
			tasks: []model.OAuth2SessionLogoutTask{unknown},
			setup: func(storage *mocks.MockStorage) {
				storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 3, 0, now.Add(time.Minute)).Return(true, nil)
				storage.EXPECT().DeleteOAuth2SessionLogoutTask(gomock.Any(), 3, 1).Return(true, nil)
			},
			processed: 1,
		},
		{
			name:  "ShouldBackOffExponentially",
			tasks: []model.OAuth2SessionLogoutTask{retried, exhausted},
			setup: func(storage *mocks.MockStorage) {
				storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 4, 3, now.Add(8*time.Minute)).Return(false, nil)
				storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 5, 100, now.Add(time.Hour)).Return(false, nil)
			},
			processed: 2,
		},
		{
			name:  "ShouldSkipTaskClaimedByAnotherAttempt",
			tasks: []model.OAuth2SessionLogoutTask{revoke},
			setup: func(storage *mocks.MockStorage) {
				storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 1, 0, now.Add(time.Minute)).Return(false, nil)
			},
			processed: 1,
		},
		{
			name:  "ShouldKeepTaskWhenRevokeFails",
			tasks: []model.OAuth2SessionLogoutTask{revoke, expired},
			setup: func(storage *mocks.MockStorage) {
				storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 1, 0, now.Add(time.Minute)).Return(true, nil)
				storage.EXPECT().RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), "sid").Return(errors.New("revoke failed"))
				storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 2, 1, now.Add(2*time.Minute)).Return(true, nil)
				storage.EXPECT().RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), "sid").Return(nil)
				storage.EXPECT().DeleteOAuth2SessionLogoutTask(gomock.Any(), 2, 2).Return(true, nil)
			},
			processed: 1,
			err:       "error processing oauth2 session logout task of kind 'revoke' with session id 'sid': revoke failed",
		},
		{
			name:  "ShouldReturnClaimError",
			tasks: []model.OAuth2SessionLogoutTask{revoke},
			setup: func(storage *mocks.MockStorage) {
				storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 1, 0, now.Add(time.Minute)).Return(false, errors.New("claim failed"))
			},
			err: "error processing oauth2 session logout task of kind 'revoke' with session id 'sid': claim failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockStorage(ctrl)

			storage.EXPECT().LoadOAuth2SessionLogoutTasksDue(gomock.Any(), now, 10).Return(tc.tasks, nil)

			if tc.setup != nil {
				tc.setup(storage)
			}

			collector := oidc.NewLogoutTaskCollector(storage, clock.NewFixed(now), nil)

			processed, err := collector.Collect(context.Background(), 10)

			assert.Equal(t, tc.processed, processed)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestLogoutTaskCollectorShouldReturnLoadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	storage.EXPECT().LoadOAuth2SessionLogoutTasksDue(gomock.Any(), gomock.Any(), 500).Return(nil, errors.New("load failed"))

	collector := oidc.NewLogoutTaskCollector(storage, clock.New(), nil)

	require.EqualError(t, collector.GarbageCollection(context.Background()), "load failed")
	assert.Equal(t, time.Minute, collector.GarbageCollectionFrequency(context.Background()))
}

func TestLogoutTaskCollectorShouldDeliverBackChannelLogout(t *testing.T) {
	now := time.Unix(1700000000, 0)

	task := model.OAuth2SessionLogoutTask{ID: 1, Kind: model.OAuth2SessionLogoutTaskKindBackChannel, SessionID: "sid", ClientID: "client", Attempts: 2, ExpiresAt: now.Add(time.Hour)}
	expired := model.OAuth2SessionLogoutTask{ID: 1, Kind: model.OAuth2SessionLogoutTaskKindBackChannel, SessionID: "sid", ClientID: "client", Attempts: 2, ExpiresAt: now}

	testCases := []struct {
		name      string
		task      model.OAuth2SessionLogoutTask
		acquire   error
		claimed   bool
		done      bool
		deliver   error
		calls     []string
		processed int
		err       string
	}{
		{
			name:      "ShouldDeleteDeliveredTask",
			task:      task,
			claimed:   true,
			done:      true,
			calls:     []string{"acquire", "claim", "deliver", "delete", "release"},
			processed: 1,
		},
		{
			name:    "ShouldKeepTaskWhichMaySucceedIfRetried",
			task:    task,
			claimed: true,
			deliver: errors.New("the client responded with status code '503'"),
			calls:   []string{"acquire", "claim", "deliver", "release"},
			err:     "error processing oauth2 session logout task of kind 'backchannel' with session id 'sid': error delivering the Logout Token to the client with id 'client': the client responded with status code '503'",
		},
		{
			name:    "ShouldDeleteTaskWhichWontSucceedIfRetried",
			task:    task,
			claimed: true,
			done:    true,
			deliver: errors.New("the client responded with status code '400'"),
			calls:   []string{"acquire", "claim", "deliver", "delete", "release"},
			err:     "error processing oauth2 session logout task of kind 'backchannel' with session id 'sid': error delivering the Logout Token to the client with id 'client': the client responded with status code '400'",
		},
		{
			name:      "ShouldSkipTaskClaimedByAnotherAttempt",
			task:      task,
			calls:     []string{"acquire", "claim", "release"},
			processed: 1,
		},
		{
			name:    "ShouldDeleteExpiredTaskWithoutDelivering",
			task:    expired,
			claimed: true,
			calls:   []string{"acquire", "claim", "delete", "release"},
			err:     "error processing oauth2 session logout task of kind 'backchannel' with session id 'sid': the Logout Token for the client with id 'client' was not delivered before the task expired",
		},
		{
			name:    "ShouldNotClaimWithoutASlot",
			task:    task,
			acquire: context.DeadlineExceeded,
			calls:   []string{"acquire"},
			err:     "error processing oauth2 session logout task of kind 'backchannel' with session id 'sid': context deadline exceeded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockStorage(ctrl)

			deliverer := &testLogoutTokenDeliverer{acquire: tc.acquire, done: tc.done, err: tc.deliver}

			storage.EXPECT().LoadOAuth2SessionLogoutTasksDue(gomock.Any(), now, 10).Return([]model.OAuth2SessionLogoutTask{tc.task}, nil)

			storage.EXPECT().
				ClaimOAuth2SessionLogoutTask(gomock.Any(), 1, 2, now.Add(4*time.Minute)).
				MaxTimes(1).
				DoAndReturn(func(_ context.Context, _, _ int, _ time.Time) (bool, error) {
					deliverer.record("claim")

					return tc.claimed, nil
				})

			storage.EXPECT().
				DeleteOAuth2SessionLogoutTask(gomock.Any(), 1, 3).
				MaxTimes(1).
				DoAndReturn(func(_ context.Context, _, _ int) (bool, error) {
					deliverer.record("delete")

					return true, nil
				})

			collector := oidc.NewLogoutTaskCollector(storage, clock.NewFixed(now), deliverer)

			processed, err := collector.Collect(context.Background(), 10)

			assert.Equal(t, tc.processed, processed)
			assert.Equal(t, tc.calls, deliverer.calls)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestLogoutTaskCollectorShouldDeliverSessions(t *testing.T) {
	now := time.Unix(1700000000, 0)

	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	deliverer := &testLogoutTokenDeliverer{done: true}

	storage.EXPECT().
		LoadOAuth2SessionLogoutTasksBySessionID(gomock.Any(), "a", model.OAuth2SessionLogoutTaskKindBackChannel).
		Return([]model.OAuth2SessionLogoutTask{{ID: 1, Kind: model.OAuth2SessionLogoutTaskKindBackChannel, SessionID: "a", ClientID: "client", ExpiresAt: now.Add(time.Hour)}}, nil)
	storage.EXPECT().
		LoadOAuth2SessionLogoutTasksBySessionID(gomock.Any(), "b", model.OAuth2SessionLogoutTaskKindBackChannel).
		Return(nil, errors.New("connection refused"))
	storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 1, 0, now.Add(time.Minute)).Return(true, nil)
	storage.EXPECT().DeleteOAuth2SessionLogoutTask(gomock.Any(), 1, 1).Return(true, nil)

	collector := oidc.NewLogoutTaskCollector(storage, clock.NewFixed(now), deliverer)

	assert.EqualError(t, collector.DeliverSessions(context.Background(), []string{"a", "b"}), "connection refused")
	assert.Equal(t, []string{"acquire", "deliver", "release"}, deliverer.calls)
}

func TestLogoutTaskCollectorShouldDeleteADeliveredTaskAfterTheDeadline(t *testing.T) {
	now := time.Unix(1700000000, 0)

	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	deliverer := &testLogoutTokenDeliverer{done: true, delay: time.Millisecond * 100}

	storage.EXPECT().
		LoadOAuth2SessionLogoutTasksBySessionID(gomock.Any(), "sid", model.OAuth2SessionLogoutTaskKindBackChannel).
		Return([]model.OAuth2SessionLogoutTask{{ID: 1, Kind: model.OAuth2SessionLogoutTaskKindBackChannel, SessionID: "sid", ClientID: "client", ExpiresAt: now.Add(time.Hour)}}, nil)
	storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 1, 0, now.Add(time.Minute)).Return(true, nil)
	storage.EXPECT().
		DeleteOAuth2SessionLogoutTask(gomock.Any(), 1, 1).
		DoAndReturn(func(ctx context.Context, _, _ int) (bool, error) {
			return ctx.Err() == nil, ctx.Err()
		})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*20)
	defer cancel()

	collector := oidc.NewLogoutTaskCollector(storage, clock.NewFixed(now), deliverer)

	assert.NoError(t, collector.DeliverSessions(ctx, []string{"sid"}))
}

func TestLogoutTaskCollectorShouldRemoveExpiredBackChannelTasksWithoutADeliverer(t *testing.T) {
	now := time.Unix(1700000000, 0)

	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)

	storage.EXPECT().
		LoadOAuth2SessionLogoutTasksDue(gomock.Any(), now, 10).
		Return([]model.OAuth2SessionLogoutTask{{ID: 1, Kind: model.OAuth2SessionLogoutTaskKindBackChannel, SessionID: "sid", ExpiresAt: now}}, nil)
	storage.EXPECT().ClaimOAuth2SessionLogoutTask(gomock.Any(), 1, 0, now.Add(time.Minute)).Return(true, nil)
	storage.EXPECT().DeleteOAuth2SessionLogoutTask(gomock.Any(), 1, 1).Return(true, nil)

	collector := oidc.NewLogoutTaskCollector(storage, clock.NewFixed(now), nil)

	processed, err := collector.Collect(context.Background(), 10)

	require.NoError(t, err)
	assert.Equal(t, 1, processed)
}

type testLogoutTokenDeliverer struct {
	mu      sync.Mutex
	calls   []string
	acquire error
	delay   time.Duration
	done    bool
	err     error
}

func (d *testLogoutTokenDeliverer) record(call string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.calls = append(d.calls, call)
}

func (d *testLogoutTokenDeliverer) AcquireBackChannelLogoutSlot(_ context.Context) (release func(), err error) {
	d.record("acquire")

	if d.acquire != nil {
		return nil, d.acquire
	}

	return func() { d.record("release") }, nil
}

func (d *testLogoutTokenDeliverer) DeliverBackChannelLogout(_ context.Context, _ model.OAuth2SessionLogoutTask) (done bool, err error) {
	d.record("deliver")

	time.Sleep(d.delay)

	return d.done, d.err
}

func TestLogoutRevokeHorizon(t *testing.T) {
	testCases := []struct {
		name     string
		config   *schema.IdentityProvidersOpenIDConnect
		expected time.Duration
	}{
		{"ShouldUseDefaultsWithoutConfiguration", nil, 11 * time.Minute},
		{"ShouldUseLongestCodeLifespan", &schema.IdentityProvidersOpenIDConnect{Lifespans: schema.IdentityProvidersOpenIDConnectLifespans{DeviceCode: 30 * time.Minute}}, 31 * time.Minute},
		{"ShouldUseLongestCustomCodeLifespan", &schema.IdentityProvidersOpenIDConnect{Lifespans: schema.IdentityProvidersOpenIDConnectLifespans{Custom: map[string]schema.IdentityProvidersOpenIDConnectLifespan{
			"custom": {Grants: schema.IdentityProvidersOpenIDConnectLifespanGrants{AuthorizeCode: schema.IdentityProvidersOpenIDConnectLifespanToken{AuthorizeCode: time.Hour}}},
		}}}, time.Hour + time.Minute},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.LogoutRevokeHorizon(tc.config))

			now := time.Unix(1700000000, 0)

			schedule := oidc.NewLogoutRevokeSchedule(tc.config, now)

			assert.Equal(t, now, schedule.CreatedAt)
			assert.Equal(t, now.Add(oidc.LogoutRevokeGrace), schedule.NextAttemptAt)
			assert.Equal(t, now.Add(tc.expected), schedule.ExpiresAt)
		})
	}
}
