// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestLogoutPOSTShouldDeliverBackChannelLogout(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	sector := &url.URL{Scheme: "https", Host: "sector.example.com"}

	userSession, issuer := setupTestBackChannelLogout(t, mock,
		schema.IdentityProvidersOpenIDConnectClient{ID: "notified", BackChannelLogoutURI: rp.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "notified-sector", BackChannelLogoutURI: rp.URL, SectorIdentifierURI: sector},
		schema.IdentityProvidersOpenIDConnectClient{ID: "no-uri"},
	)

	subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", testUsername)
	require.NoError(t, err)

	subjectSector, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, sector.String(), testUsername)
	require.NoError(t, err)

	sid, sidSector := uuid.Must(uuid.NewRandom()).String(), uuid.Must(uuid.NewRandom()).String()

	store := setupTestBackChannelLogoutStore(mock, issuer, userSession.PublicID,
		testBackChannelLogoutParticipant{sid: sid, clientID: "notified", subject: subject.String()},
		testBackChannelLogoutParticipant{sid: sidSector, clientID: "notified-sector", subject: subjectSector.String()},
		testBackChannelLogoutParticipant{sid: sid, clientID: "no-uri", subject: subject.String()},
		testBackChannelLogoutParticipant{sid: sid, clientID: "removed", subject: subject.String()},
	)

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())

	received := rp.Received()

	require.Len(t, received, 2)

	byAudience := map[string]map[string]any{}

	for _, claims := range received {
		aud, ok := claims[oidc.ClaimAudience].([]any)
		require.True(t, ok)
		require.Len(t, aud, 1)

		byAudience[aud[0].(string)] = claims
	}

	require.Contains(t, byAudience, "notified")
	require.Contains(t, byAudience, "notified-sector")

	issuerURL, err := mock.Ctx.IssuerURL()
	require.NoError(t, err)

	for _, request := range rp.Requests() {
		aud, _ := request.claims[oidc.ClaimAudience].([]any)

		require.Len(t, aud, 1)

		assertTestBackChannelLogoutRequest(t, request, issuerURL.String(), aud[0].(string))
	}

	assert.NotEqual(t, byAudience["notified"][oidc.ClaimJWTID], byAudience["notified-sector"][oidc.ClaimJWTID])

	assert.Equal(t, sid, byAudience["notified"][oidc.ClaimSessionID])
	assert.Equal(t, subject.String(), byAudience["notified"][oidc.ClaimSubject])
	assert.Equal(t, sidSector, byAudience["notified-sector"][oidc.ClaimSessionID])
	assert.Equal(t, subjectSector.String(), byAudience["notified-sector"][oidc.ClaimSubject])

	schedule := store.Schedule()

	require.NotNil(t, schedule.BackChannel)
	assert.Equal(t, issuerURL.String(), schedule.BackChannel.IssuerURL)
	assert.Equal(t, schedule.CreatedAt.Add(time.Minute), schedule.BackChannel.NextAttemptAt)
	assert.Equal(t, schedule.CreatedAt.Add(24*time.Hour), schedule.BackChannel.ExpiresAt)
	assert.Empty(t, store.Remaining())
}

func TestLogoutPOSTShouldGiveAQueuedDeliveryItsFullBudget(t *testing.T) {
	testCases := []struct {
		name    string
		hold    time.Duration
		delay   time.Duration
		timeout time.Duration
	}{
		{name: "ShouldNotSpendTheBudgetWaitingForASlot", hold: time.Millisecond * 2500},
		{name: "ShouldNotBoundTheDeliveryByTheInitialDeadline", hold: time.Millisecond * 800, delay: time.Millisecond * 500, timeout: time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.timeout != 0 {
				timeout := oidcBackChannelLogoutDeliveryTimeout
				oidcBackChannelLogoutDeliveryTimeout = tc.timeout

				t.Cleanup(func() { oidcBackChannelLogoutDeliveryTimeout = timeout })
			}

			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			rp := newTestBackChannelLogoutRP(t)
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				time.Sleep(tc.delay)

				rp.Config.Handler.ServeHTTP(rw, r)
			}))

			t.Cleanup(server.Close)

			userSession, issuer := setupTestBackChannelLogout(t, mock,
				schema.IdentityProvidersOpenIDConnectClient{ID: "notified", BackChannelLogoutURI: server.URL},
			)

			setTestBackChannelLogoutHTTPClientWithoutRetries(mock)

			mock.Ctx.Providers.OpenIDConnect.BackChannelLogoutLifespan = time.Second * 2
			mock.Ctx.Providers.OpenIDConnect.BackChannelLogoutConcurrency = 1

			release, err := mock.Ctx.Providers.OpenIDConnect.AcquireBackChannelLogoutSlot(t.Context())
			require.NoError(t, err)

			time.AfterFunc(tc.hold, release)

			sid := uuid.Must(uuid.NewRandom()).String()

			store := setupTestBackChannelLogoutStore(mock, issuer, userSession.PublicID, testBackChannelLogoutParticipant{sid: sid, clientID: "notified"})

			mock.Ctx.Request.SetBodyString(`{}`)

			LogoutPOST(mock.Ctx)

			received := rp.Received()

			require.Len(t, received, 1)
			assert.Equal(t, sid, received[0][oidc.ClaimSessionID])
			assert.Empty(t, store.Remaining())
		})
	}
}

func TestLogoutPOSTShouldNotWaitForBackChannelLogoutDelivery(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	release := make(chan struct{})

	rp := newTestBackChannelLogoutRP(t)
	blocking := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		<-release

		rp.Config.Handler.ServeHTTP(rw, r)
	}))

	t.Cleanup(blocking.Close)

	userSession, issuer := setupTestBackChannelLogout(t, mock,
		schema.IdentityProvidersOpenIDConnectClient{ID: "notified", BackChannelLogoutURI: blocking.URL},
	)

	sid := uuid.Must(uuid.NewRandom()).String()

	setupTestBackChannelLogoutStore(mock, issuer, userSession.PublicID, testBackChannelLogoutParticipant{sid: sid, clientID: "notified"})

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())

	rp.mu.Lock()
	assert.Empty(t, rp.requests)
	rp.mu.Unlock()

	close(release)

	received := rp.Received()

	require.Len(t, received, 1)
	assert.Equal(t, sid, received[0][oidc.ClaimSessionID])
}

func TestLogoutPOSTShouldLimitBackChannelLogoutConcurrencyAcrossSessions(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	var (
		mu                 sync.Mutex
		inflight, observed int
	)

	rp := newTestBackChannelLogoutRP(t)
	slow := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		mu.Lock()
		inflight++
		observed = max(observed, inflight)
		mu.Unlock()

		time.Sleep(time.Millisecond * 50)

		mu.Lock()
		inflight--
		mu.Unlock()

		rp.Config.Handler.ServeHTTP(rw, r)
	}))

	t.Cleanup(slow.Close)

	sector := &url.URL{Scheme: "https", Host: "sector.example.com"}

	userSession, issuer := setupTestBackChannelLogout(t, mock,
		schema.IdentityProvidersOpenIDConnectClient{ID: "a", BackChannelLogoutURI: slow.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "b", BackChannelLogoutURI: slow.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "c", BackChannelLogoutURI: slow.URL, SectorIdentifierURI: sector},
	)

	mock.Ctx.Providers.OpenIDConnect.BackChannelLogoutConcurrency = 1

	sid, sidSector := uuid.Must(uuid.NewRandom()).String(), uuid.Must(uuid.NewRandom()).String()

	setupTestBackChannelLogoutStore(mock, issuer, userSession.PublicID,
		testBackChannelLogoutParticipant{sid: sid, clientID: "a"},
		testBackChannelLogoutParticipant{sid: sid, clientID: "b"},
		testBackChannelLogoutParticipant{sid: sidSector, clientID: "c"},
	)

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	assert.Len(t, rp.Received(), 3)

	mu.Lock()
	assert.Equal(t, 1, observed)
	mu.Unlock()
}

func TestLogoutPOSTShouldDeliverBackChannelLogoutToASingleClient(t *testing.T) {
	testCases := []struct {
		name      string
		status    int
		received  int
		remaining int
	}{
		{"ShouldDeliverWithoutTheSubjectWhenItWasNotRecorded", http.StatusOK, 1, 0},
		{"ShouldLeaveServerErrorsForTheWorker", http.StatusServiceUnavailable, 0, 1},
		{"ShouldLeaveRateLimitingForTheWorker", http.StatusTooManyRequests, 0, 1},
		{"ShouldNotRetryClientErrors", http.StatusBadRequest, 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			var attempts atomic.Int32

			rp := newTestBackChannelLogoutRP(t)
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				attempts.Add(1)

				if tc.status != http.StatusOK {
					rw.WriteHeader(tc.status)

					return
				}

				rp.Config.Handler.ServeHTTP(rw, r)
			}))

			t.Cleanup(server.Close)

			userSession, issuer := setupTestBackChannelLogout(t, mock,
				schema.IdentityProvidersOpenIDConnectClient{ID: "notified", BackChannelLogoutURI: server.URL},
			)

			setTestBackChannelLogoutHTTPClientWithoutRetries(mock)

			sid := uuid.Must(uuid.NewRandom()).String()

			store := setupTestBackChannelLogoutStore(mock, issuer, userSession.PublicID, testBackChannelLogoutParticipant{sid: sid, clientID: "notified"})

			mock.Ctx.Request.SetBodyString(`{}`)

			LogoutPOST(mock.Ctx)

			remaining := store.Remaining()
			received := rp.Received()

			assert.Equal(t, int32(1), attempts.Load())
			require.Len(t, received, tc.received)
			require.Len(t, remaining, tc.remaining)

			for _, claims := range received {
				assert.Equal(t, sid, claims[oidc.ClaimSessionID])
				assert.NotContains(t, claims, oidc.ClaimSubject)
			}

			for _, task := range remaining {
				assert.Equal(t, 1, task.Attempts)
				assert.WithinDuration(t, mock.Ctx.GetClock().Now().Add(time.Minute), task.NextAttemptAt, time.Second)
			}
		})
	}
}

type testBackChannelLogoutParticipant struct {
	sid      string
	clientID string
	subject  string
}

type testBackChannelLogoutStore struct {
	mu         sync.Mutex
	failSaves  int
	failRevoke bool
	next       int
	schedule   model.OAuth2SessionLogoutTaskSchedule
	tasks      map[int]model.OAuth2SessionLogoutTask
}

func setupTestBackChannelLogoutStore(mock *mocks.MockAutheliaCtx, issuer, publicID string, participants ...testBackChannelLogoutParticipant) (store *testBackChannelLogoutStore) {
	store = &testBackChannelLogoutStore{tasks: map[int]model.OAuth2SessionLogoutTask{}}

	mock.StorageMock.EXPECT().
		EndOAuth2Session(gomock.Any(), issuer, publicID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, schedule model.OAuth2SessionLogoutTaskSchedule) ([]string, error) {
			store.mu.Lock()
			defer store.mu.Unlock()

			store.schedule = schedule

			var sids []string

			for _, participant := range participants {
				if !slices.Contains(sids, participant.sid) {
					sids = append(sids, participant.sid)
				}

				if bc := schedule.BackChannel; bc != nil {
					store.add(model.OAuth2SessionLogoutTask{
						Kind:          model.OAuth2SessionLogoutTaskKindBackChannel,
						SessionID:     participant.sid,
						ClientID:      participant.clientID,
						Subject:       participant.subject,
						IssuerURL:     bc.IssuerURL,
						NextAttemptAt: bc.NextAttemptAt,
						CreatedAt:     schedule.CreatedAt,
						ExpiresAt:     bc.ExpiresAt,
					})
				}
			}

			return sids, nil
		})

	setupTestBackChannelLogoutTasks(mock, store)

	return store
}

func setupTestBackChannelLogoutTasks(mock *mocks.MockAutheliaCtx, store *testBackChannelLogoutStore) {
	mock.StorageMock.EXPECT().
		RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, _ string) error {
			store.mu.Lock()
			defer store.mu.Unlock()

			if store.failRevoke {
				return errors.New("connection refused")
			}

			return nil
		})

	mock.StorageMock.EXPECT().
		SaveOAuth2SessionLogoutTask(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, task model.OAuth2SessionLogoutTask) error {
			store.mu.Lock()
			defer store.mu.Unlock()

			if store.failSaves > 0 {
				store.failSaves--

				return errors.New("connection refused")
			}

			store.add(task)

			return nil
		})

	mock.StorageMock.EXPECT().
		LoadOAuth2SessionLogoutTasksBySessionID(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, sid, kind string) (tasks []model.OAuth2SessionLogoutTask, err error) {
			store.mu.Lock()
			defer store.mu.Unlock()

			for id := range store.next {
				if task, ok := store.tasks[id]; ok && task.SessionID == sid && task.Kind == kind {
					tasks = append(tasks, task)
				}
			}

			return tasks, nil
		})

	mock.StorageMock.EXPECT().
		ClaimOAuth2SessionLogoutTask(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, id, attempts int, next time.Time) (bool, error) {
			store.mu.Lock()
			defer store.mu.Unlock()

			task, ok := store.tasks[id]
			if !ok || task.Attempts != attempts {
				return false, nil
			}

			task.Attempts, task.NextAttemptAt = attempts+1, next
			store.tasks[id] = task

			return true, nil
		})

	mock.StorageMock.EXPECT().
		DeleteOAuth2SessionLogoutTask(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, id, attempts int) (bool, error) {
			store.mu.Lock()
			defer store.mu.Unlock()

			if task, ok := store.tasks[id]; !ok || task.Attempts != attempts {
				return false, nil
			}

			delete(store.tasks, id)

			return true, nil
		})
}

func (s *testBackChannelLogoutStore) add(task model.OAuth2SessionLogoutTask) {
	task.ID = s.next
	s.tasks[s.next] = task
	s.next++
}

func (s *testBackChannelLogoutStore) Schedule() model.OAuth2SessionLogoutTaskSchedule {
	oidcBackChannelLogoutDeliveries.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.schedule
}

func (s *testBackChannelLogoutStore) All() (tasks []model.OAuth2SessionLogoutTask) {
	oidcBackChannelLogoutDeliveries.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	for id := range s.next {
		if task, ok := s.tasks[id]; ok {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func (s *testBackChannelLogoutStore) Remaining() (tasks []model.OAuth2SessionLogoutTask) {
	oidcBackChannelLogoutDeliveries.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	for id := range s.next {
		if task, ok := s.tasks[id]; ok && task.Kind == model.OAuth2SessionLogoutTaskKindBackChannel {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

type testBackChannelLogoutRequest struct {
	method      string
	contentType string
	header      map[string]any
	claims      map[string]any
}

type testBackChannelLogoutRP struct {
	*httptest.Server

	mu       sync.Mutex
	requests []testBackChannelLogoutRequest
}

func newTestBackChannelLogoutRP(t *testing.T) (rp *testBackChannelLogoutRP) {
	t.Helper()

	rp = &testBackChannelLogoutRP{}

	rp.Server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			rw.WriteHeader(http.StatusBadRequest)

			return
		}

		parts := strings.Split(r.PostFormValue("logout_token"), ".")
		if len(parts) != 3 {
			rw.WriteHeader(http.StatusBadRequest)

			return
		}

		request := testBackChannelLogoutRequest{method: r.Method, contentType: r.Header.Get("Content-Type")}

		for i, segment := range []*map[string]any{&request.header, &request.claims} {
			decoded, err := base64.RawURLEncoding.DecodeString(parts[i])
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)

				return
			}

			*segment = map[string]any{}

			if err = json.Unmarshal(decoded, segment); err != nil {
				rw.WriteHeader(http.StatusBadRequest)

				return
			}
		}

		rp.mu.Lock()
		rp.requests = append(rp.requests, request)
		rp.mu.Unlock()

		rw.Header().Set("Cache-Control", "no-store")
		rw.WriteHeader(http.StatusOK)
	}))

	t.Cleanup(rp.Close)

	return rp
}

func (rp *testBackChannelLogoutRP) Requests() []testBackChannelLogoutRequest {
	oidcBackChannelLogoutDeliveries.Wait()

	rp.mu.Lock()
	defer rp.mu.Unlock()

	return append([]testBackChannelLogoutRequest(nil), rp.requests...)
}

func (rp *testBackChannelLogoutRP) Received() []map[string]any {
	requests := rp.Requests()

	claims := make([]map[string]any, len(requests))

	for i, request := range requests {
		claims[i] = request.claims
	}

	return claims
}

func assertTestBackChannelLogoutRequest(t *testing.T, request testBackChannelLogoutRequest, issuer, clientID string) {
	t.Helper()

	assert.Equal(t, http.MethodPost, request.method)
	assert.Equal(t, "application/x-www-form-urlencoded", request.contentType)

	assert.NotEqual(t, "none", request.header["alg"])
	assert.Equal(t, "logout+jwt", request.header["typ"])

	assert.Equal(t, issuer, request.claims[oidc.ClaimIssuer])
	assert.Equal(t, []any{clientID}, request.claims[oidc.ClaimAudience])
	assert.NotEmpty(t, request.claims[oidc.ClaimIssuedAt])
	assert.NotEmpty(t, request.claims[oidc.ClaimExpirationTime])
	assert.NotEmpty(t, request.claims[oidc.ClaimJWTID])

	events, ok := request.claims["events"].(map[string]any)

	require.True(t, ok, "the 'events' claim must be present")
	assert.Equal(t, map[string]any{}, events["http://schemas.openid.net/event/backchannel-logout"])

	assert.True(t, request.claims[oidc.ClaimSubject] != nil || request.claims[oidc.ClaimSessionID] != nil)
	assert.NotContains(t, request.claims, oidc.ClaimNonce)
}

func setupTestBackChannelLogout(t *testing.T, mock *mocks.MockAutheliaCtx, clients ...schema.IdentityProvidersOpenIDConnectClient) (userSession session.UserSession, issuer string) {
	t.Helper()

	setupTestOIDCLogoutProvider(t, mock, clients...)

	return mustSaveTestUserSession(t, mock, testUsername)
}

func setTestBackChannelLogoutHTTPClientWithoutRetries(mock *mocks.MockAutheliaCtx) {
	client := oidc.NewHTTPClient(nil)
	client.RetryMax = 0

	mock.Ctx.Providers.OpenIDConnect.HTTPClient = client
}

func newTestOIDCLoopbackNetworks(t *testing.T) (networks []*net.IPNet) {
	t.Helper()

	for _, cidr := range []string{"127.0.0.0/8", "::1/128"} {
		_, network, err := net.ParseCIDR(cidr)
		require.NoError(t, err)

		networks = append(networks, network)
	}

	return networks
}
