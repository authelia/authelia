// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/appauthenticator"
	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/storage"
)

const appTestTargetURL = "https://two-factor.example.com/"

type appChallengeData struct {
	Number      any   `json:"number"`
	NonceLength int   `json:"nonce_length"`
	ExpiresAt   int64 `json:"expires_at"`
}

func appSignBody() map[string]any {
	return map[string]any{"targetURL": appTestTargetURL}
}

func (s *appPortalSession) challenge() string {
	s.p.t.Helper()

	data := appChallengeData{}

	appPortalOK(s.p.t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorPOST), &data)

	number, ok := data.Number.(string)
	require.True(s.p.t, ok, "number must be a JSON string")

	return number
}

func (s *appPortalSession) status() (*middlewares.AutheliaCtx, map[string]any) {
	s.p.t.Helper()

	ctx := s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorStatusPOST)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		return ctx, nil
	}

	response := appPortalResponse{}

	require.NoError(s.p.t, json.Unmarshal(ctx.Response.Body(), &response))

	if response.Status != "OK" {
		return ctx, nil
	}

	data := map[string]any{}

	if len(response.Data) != 0 {
		require.NoError(s.p.t, json.Unmarshal(response.Data, &data))
	}

	return ctx, data
}

func (s *appPortalSession) state() string {
	s.p.t.Helper()

	ctx, data := s.status()

	require.NotNil(s.p.t, data, string(ctx.Response.Body()))

	state, ok := data["state"].(string)
	require.True(s.p.t, ok, string(ctx.Response.Body()))

	return state
}

func (p *appPortal) expectDetails() {
	username := "john"

	p.mock.UserProviderMock.EXPECT().GetDetails(username).Return(&authentication.UserDetails{Username: username, DisplayName: username, Emails: []string{username + "@example.com"}}, nil)
}

func appAttempts(p *appPortal) []model.AuthenticationAttempt {
	var attempts []model.AuthenticationAttempt

	for _, attempt := range p.env.logs.Attempts() {
		if attempt.Type == regulation.AuthTypeApp {
			attempts = append(attempts, attempt)
		}
	}

	return attempts
}

func TestAppAuthenticatorShouldApproveAndUpgradeSession(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	d := s.registerDevice()

	now := p.env.clock.Now()

	data := appChallengeData{}

	appPortalOK(t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorPOST), &data)

	number, ok := data.Number.(string)
	require.True(t, ok)
	assert.Regexp(t, `^[0-9]{2}$`, number)
	assert.Equal(t, 2, data.NonceLength)
	assert.Equal(t, now.Add(2*time.Minute).Unix(), data.ExpiresAt)

	assert.Equal(t, "pending", s.state())

	pulled, err := d.Check()
	require.NoError(t, err)
	require.NotNil(t, pulled)
	assert.Equal(t, map[string]any{
		"portal_url":    appPortalBaseURL,
		"requested_url": appTestTargetURL,
		"remote_ip":     "198.51.100.4",
		"user_agent":    "Firefox on macOS",
		"requested_at":  float64(now.Unix()),
	}, pulled.Details)

	result, err := d.Approve(number)
	require.NoError(t, err)
	assert.Equal(t, "approved", result)

	before := s.cookie

	p.expectDetails()

	ctx, body := s.status()
	require.NotNil(t, body, string(ctx.Response.Body()))
	assert.Equal(t, map[string]any{"redirect": appTestTargetURL}, body)
	assert.NotEqual(t, before, s.cookie)

	userSession := s.session()
	assert.Equal(t, authentication.TwoFactor, userSession.AuthenticationLevel(false))
	assert.True(t, userSession.AuthenticationMethodRefs.App)
	assert.ElementsMatch(t, []string{"pwd", "kba", "pop", "swk", "mfa", "mca"}, userSession.AuthenticationMethodRefs.MarshalRFC8176())
	assert.Equal(t, now.Unix(), userSession.SecondFactorAuthnTimestamp)

	row, err := p.env.storage.LoadAppAuthenticatorChallengeByPortalBinding(t.Context(), s.binding())
	require.NoError(t, err)
	assert.True(t, row.ConsumedAt.Valid)

	appPortalKO(t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorStatusPOST), fasthttp.StatusUnauthorized)

	attempts := appAttempts(p)
	require.Len(t, attempts, 1)
	assert.True(t, attempts[0].Successful)
	assert.Equal(t, "john", attempts[0].Username)
}

func TestAppAuthenticatorShouldRespondWithoutTarget(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	d := s.registerDevice()

	appPortalOK(t, s.do(fasthttp.MethodPost, nil, handlers.AppAuthenticatorPOST), nil)

	_, err := d.Check()
	require.NoError(t, err)
	require.NotNil(t, d.Last)
	assert.NotContains(t, d.Last.Details, "requested_url")

	row, err := p.env.storage.LoadAppAuthenticatorChallengeByPortalBinding(t.Context(), s.binding())
	require.NoError(t, err)

	result, err := d.Approve(row.Number)
	require.NoError(t, err)
	assert.Equal(t, "approved", result)

	p.expectDetails()

	ctx := s.do(fasthttp.MethodPost, nil, handlers.AppAuthenticatorStatusPOST)
	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.JSONEq(t, `{"status":"OK","data":{"redirect":"https://www.example.com"}}`, string(ctx.Response.Body()))
}

func TestAppAuthenticatorShouldIgnoreUnsafeTarget(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	d := s.registerDevice()

	appPortalOK(t, s.do(fasthttp.MethodPost, map[string]any{"targetURL": "https://evil.example.org/"}, handlers.AppAuthenticatorPOST), nil)

	pulled, err := d.Check()
	require.NoError(t, err)
	require.NotNil(t, pulled)
	assert.NotContains(t, pulled.Details, "requested_url")
}

func TestAppAuthenticatorStatusShouldNotUpgradeOtherSession(t *testing.T) {
	p := newAppPortal(t)
	owner := p.login("john")
	other := p.login("john")
	d := owner.registerDevice()

	number := owner.challenge()

	appPortalKO(t, other.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorStatusPOST), fasthttp.StatusUnauthorized)

	other.challenge()

	stale := &appPortalSession{p: p, cookie: owner.cookie}

	owner.do(fasthttp.MethodGet, nil, func(ctx *middlewares.AutheliaCtx) {
		require.NoError(t, ctx.RegenerateSession())
	})

	assert.NotEqual(t, stale.cookie, owner.cookie)

	_, err := d.Check()
	require.NoError(t, err)
	require.NotNil(t, d.Last)

	result, err := d.Approve(number)
	require.NoError(t, err)
	assert.Equal(t, "approved", result)

	assert.Equal(t, "pending", other.state())

	appPortalKO(t, stale.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorStatusPOST), fasthttp.StatusUnauthorized)

	p.expectDetails()

	_, body := owner.status()
	assert.Equal(t, map[string]any{"redirect": appTestTargetURL}, body)

	userSession := owner.session()
	assert.Equal(t, authentication.TwoFactor, userSession.AuthenticationLevel(false))

	otherSession := other.session()
	assert.Equal(t, authentication.OneFactor, otherSession.AuthenticationLevel(false))
}

func TestAppAuthenticatorShouldReportDecisions(t *testing.T) {
	t.Run("ShouldReportDenied", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")
		d := s.registerDevice()

		s.challenge()

		_, err := d.Check()
		require.NoError(t, err)

		result, err := d.Deny()
		require.NoError(t, err)
		assert.Equal(t, "denied", result)

		assert.Equal(t, "denied", s.state())

		userSession := s.session()
		assert.Equal(t, authentication.OneFactor, userSession.AuthenticationLevel(false))

		attempts := appAttempts(p)
		require.Len(t, attempts, 1)
		assert.False(t, attempts[0].Successful)
	})

	t.Run("ShouldReportRejected", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")
		d := s.registerDevice()

		number := s.challenge()

		_, err := d.Check()
		require.NoError(t, err)

		result, err := d.Approve(wrongNumber(number))
		require.NoError(t, err)
		assert.Equal(t, "rejected", result)

		assert.Equal(t, "rejected", s.state())

		attempts := appAttempts(p)
		require.Len(t, attempts, 1)
		assert.False(t, attempts[0].Successful)
	})

	t.Run("ShouldReportExpired", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")
		s.registerDevice()

		s.challenge()

		p.env.clock.Advance(2*time.Minute + time.Second)

		assert.Equal(t, "expired", s.state())
	})

	t.Run("ShouldReportExpiredWhenSettled", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")
		d := s.registerDevice()

		number := s.challenge()

		_, err := d.Check()
		require.NoError(t, err)

		p.env.clock.Advance(2*time.Minute + time.Second)

		result, err := d.Approve(number)
		require.NoError(t, err)
		assert.Equal(t, "rejected", result)

		assert.Equal(t, "expired", s.state())
	})

	t.Run("ShouldVoidPreviousChallenge", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")
		s.registerDevice()

		s.challenge()

		first, err := p.env.storage.LoadAppAuthenticatorChallengeByPortalBinding(t.Context(), s.binding())
		require.NoError(t, err)

		s.challenge()

		row, err := p.env.storage.LoadAppAuthenticatorChallengeByConfirmationID(t.Context(), first.ConfirmationID)
		require.NoError(t, err)
		assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, row.State)

		assert.Equal(t, "pending", s.state())
	})
}

func TestAppAuthenticatorShouldUpgradeWithAttestedDevice(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")

	deviceID, err := appauthenticator.NewID()
	require.NoError(t, err)

	require.NoError(t, p.env.storage.SaveAppAuthenticatorDevice(t.Context(), model.AppAuthenticatorDevice{
		CreatedAt:          p.env.clock.Now(),
		DeviceID:           deviceID,
		Username:           "john",
		BaseURL:            appPortalBaseURL,
		Name:               "Pixel",
		Model:              "Pixel 9",
		Platform:           "Android",
		ApprovalPublicKey:  []byte("approval"),
		TransportPublicKey: []byte("transport"),
		AttestationFormat:  appauthenticator.AttestationFormatAndroidKey,
		PullIDCurrent:      appauthenticator.HashID("pull"),
	}))

	s.challenge()

	row, err := p.env.storage.LoadAppAuthenticatorChallengeByPortalBinding(t.Context(), s.binding())
	require.NoError(t, err)

	row.BoundDeviceID = sql.NullString{String: deviceID, Valid: true}
	row.BoundCodeChallenge = sql.NullString{String: "code", Valid: true}

	updated, err := p.env.storage.UpdateAppAuthenticatorChallenge(t.Context(), *row, model.AppAuthenticatorChallengeStatePending)
	require.NoError(t, err)
	require.True(t, updated)

	row.State = model.AppAuthenticatorChallengeStateApproved
	row.DecidedAt = sql.NullTime{Time: p.env.clock.Now(), Valid: true}
	row.DecidedDeviceID = sql.NullString{String: deviceID, Valid: true}

	updated, err = p.env.storage.UpdateAppAuthenticatorChallengeDecision(t.Context(), *row)
	require.NoError(t, err)
	require.True(t, updated)

	p.expectDetails()

	_, body := s.status()
	assert.Equal(t, map[string]any{"redirect": appTestTargetURL}, body)

	userSession := s.session()
	assert.True(t, userSession.AuthenticationMethodRefs.AppHardware)
	assert.True(t, userSession.AuthenticationMethodRefs.AppUserVerified)
	assert.ElementsMatch(t, []string{"pwd", "kba", "pop", "hwk", "user", "mfa", "mca"}, userSession.AuthenticationMethodRefs.MarshalRFC8176())
}

func TestAppAuthenticatorShouldFailWithoutDevices(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")

	appPortalKO(t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorPOST), fasthttp.StatusUnauthorized)

	_, err := p.env.storage.LoadAppAuthenticatorChallengePending(t.Context(), "john", p.env.clock.Now())
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestAppAuthenticatorShouldFailWhenBanned(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	s.registerDevice()

	require.NoError(t, p.env.storage.SaveBannedUser(t.Context(), &model.BannedUser{
		Time:     time.Now(),
		Expires:  sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true},
		Username: "john",
		Source:   "regulation",
	}))

	appPortalKO(t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorPOST), fasthttp.StatusUnauthorized)

	_, err := p.env.storage.LoadAppAuthenticatorChallengePending(t.Context(), "john", p.env.clock.Now())
	assert.ErrorIs(t, err, sql.ErrNoRows)

	attempts := appAttempts(p)
	require.Len(t, attempts, 1)
	assert.False(t, attempts[0].Successful)
	assert.True(t, attempts[0].Banned)
}

func TestAppAuthenticatorStatusShouldFailWithoutChallenge(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")

	appPortalKO(t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorStatusPOST), fasthttp.StatusUnauthorized)

	s.registrationPUT()

	appPortalKO(t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorStatusPOST), fasthttp.StatusUnauthorized)

	anonymous := &appPortalSession{p: p}

	appPortalKO(t, anonymous.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorPOST), fasthttp.StatusUnauthorized)
	appPortalKO(t, anonymous.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorStatusPOST), fasthttp.StatusUnauthorized)
}

func TestAppAuthenticatorShouldRejectMalformedBody(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	s.registerDevice()

	appPortalKO(t, s.do(fasthttp.MethodPost, "not an object", handlers.AppAuthenticatorPOST), fasthttp.StatusUnauthorized)
	appPortalKO(t, s.do(fasthttp.MethodPost, "not an object", handlers.AppAuthenticatorStatusPOST), fasthttp.StatusUnauthorized)

	_, err := p.env.storage.LoadAppAuthenticatorChallengePending(t.Context(), "john", p.env.clock.Now())
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func (s *appPortalSession) approved() *appdeviceApproval {
	s.p.t.Helper()

	d := s.registerDevice()

	number := s.challenge()

	_, err := d.Check()
	require.NoError(s.p.t, err)

	result, err := d.Approve(number)
	require.NoError(s.p.t, err)
	require.Equal(s.p.t, "approved", result)

	row, err := s.p.env.storage.LoadAppAuthenticatorChallengeByPortalBinding(s.p.t.Context(), s.binding())
	require.NoError(s.p.t, err)

	return &appdeviceApproval{deviceID: d.DeviceID, confirmationID: row.ConfirmationID}
}

type appdeviceApproval struct {
	deviceID       string
	confirmationID string
}

func TestAppAuthenticatorStatusShouldBoundConsumption(t *testing.T) {
	t.Run("ShouldUpgradeWithinGrace", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")
		s.approved()

		p.env.clock.Advance(2*time.Minute + 59*time.Second)

		p.expectDetails()

		_, body := s.status()
		assert.Equal(t, map[string]any{"redirect": appTestTargetURL}, body)

		userSession := s.session()
		assert.Equal(t, authentication.TwoFactor, userSession.AuthenticationLevel(false))
	})

	t.Run("ShouldReportExpiredAfterGrace", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")
		approval := s.approved()

		p.env.clock.Advance(3*time.Minute + time.Second)

		assert.Equal(t, "expired", s.state())

		userSession := s.session()
		assert.Equal(t, authentication.OneFactor, userSession.AuthenticationLevel(false))

		row, err := p.env.storage.LoadAppAuthenticatorChallengeByConfirmationID(t.Context(), approval.confirmationID)
		require.NoError(t, err)
		assert.False(t, row.ConsumedAt.Valid)
	})
}

func TestAppAuthenticatorStatusShouldNotUpgradeWhenDeviceRemoved(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	approval := s.approved()

	require.NoError(t, p.env.storage.DeleteAppAuthenticatorDevice(t.Context(), approval.deviceID))

	appPortalKO(t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorStatusPOST), fasthttp.StatusUnauthorized)

	userSession := s.session()
	assert.Equal(t, authentication.OneFactor, userSession.AuthenticationLevel(false))

	row, err := p.env.storage.LoadAppAuthenticatorChallengeByConfirmationID(t.Context(), approval.confirmationID)
	require.NoError(t, err)
	assert.False(t, row.ConsumedAt.Valid)
}

func TestAppAuthenticatorStatusShouldNotConsumeWhenDetailsFail(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	approval := s.approved()

	p.mock.UserProviderMock.EXPECT().GetDetails("john").Return(nil, errors.New("backend down"))

	appPortalKO(t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorStatusPOST), fasthttp.StatusUnauthorized)

	row, err := p.env.storage.LoadAppAuthenticatorChallengeByConfirmationID(t.Context(), approval.confirmationID)
	require.NoError(t, err)
	assert.False(t, row.ConsumedAt.Valid)

	p.expectDetails()

	_, body := s.status()
	assert.Equal(t, map[string]any{"redirect": appTestTargetURL}, body)
}

type appTestFailingSaveStore struct {
	storage.Provider
}

func (s appTestFailingSaveStore) SaveAppAuthenticatorChallenge(_ context.Context, _ model.AppAuthenticatorChallenge) error {
	return errors.New("save failed")
}

func TestAppAuthenticatorShouldKeepOnePendingChallengePerBinding(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	s.registerDevice()

	ids := make([]string, 0, 3)

	for range 3 {
		s.challenge()

		row, err := p.env.storage.LoadAppAuthenticatorChallengeByPortalBinding(t.Context(), s.binding())
		require.NoError(t, err)

		ids = append(ids, row.ConfirmationID)
	}

	pending := 0

	for _, id := range ids {
		row, err := p.env.storage.LoadAppAuthenticatorChallengeByConfirmationID(t.Context(), id)
		require.NoError(t, err)

		if row.State == model.AppAuthenticatorChallengeStatePending {
			pending++

			assert.Equal(t, ids[len(ids)-1], id)
		}
	}

	assert.Equal(t, 1, pending)
}

func TestAppAuthenticatorShouldNotVoidPreviousChallengeWhenSaveFails(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	s.registerDevice()

	s.challenge()

	first, err := p.env.storage.LoadAppAuthenticatorChallengeByPortalBinding(t.Context(), s.binding())
	require.NoError(t, err)

	p.mock.Ctx.Providers.StorageProvider = appTestFailingSaveStore{Provider: p.env.storage}

	appPortalKO(t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorPOST), fasthttp.StatusUnauthorized)

	row, err := p.env.storage.LoadAppAuthenticatorChallengeByConfirmationID(t.Context(), first.ConfirmationID)
	require.NoError(t, err)
	assert.Equal(t, model.AppAuthenticatorChallengeStatePending, row.State)
}

func TestAppAuthenticatorShouldLogChallengeCreated(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	s.registerDevice()

	p.mock.Hook.Reset()

	number := s.challenge()

	var entry *logrus.Entry

	for _, e := range p.mock.Hook.AllEntries() {
		if e.Message == "App authenticator challenge created" {
			entry = e
		}
	}

	require.NotNil(t, entry)
	assert.Equal(t, logrus.InfoLevel, entry.Level)
	assert.Equal(t, "john", entry.Data["username"])
	assert.Equal(t, appPortalBaseURL, entry.Data["portal_url"])
	assert.Equal(t, appTestTargetURL, entry.Data["requested_url"])

	for _, v := range entry.Data {
		assert.NotEqual(t, number, v)
	}
}
