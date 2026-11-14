// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func testAppDevice(deviceID, username string) model.AppAuthenticatorDevice {
	return model.AppAuthenticatorDevice{
		CreatedAt:           time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		DeviceID:            deviceID,
		Username:            username,
		BaseURL:             "https://auth.example.com",
		Name:                "Jane's iPhone",
		Model:               "iPhone15,2",
		Platform:            "ios",
		ApprovalPublicKey:   []byte{1, 2, 3},
		TransportPublicKey:  []byte{4, 5, 6},
		AttestationFormat:   "apple-appattest",
		AttestationIdentity: sql.NullString{String: "TEAM.bundle", Valid: true},
		AttestationDetail:   sql.NullString{String: "production", Valid: true},
		AttestationKeyID:    []byte{7, 8},
		PullIDCurrent:       []byte{9, 9, 9},
		PullIDPrevious:      []byte{8, 8, 8},
	}
}

func testAppChallenge(confirmationID, username string, binding []byte, expires time.Time) model.AppAuthenticatorChallenge {
	return model.AppAuthenticatorChallenge{
		CreatedAt:      time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		ExpiresAt:      expires,
		ConfirmationID: confirmationID,
		Username:       username,
		PortalBinding:  binding,
		Challenge:      "challenge",
		Number:         "0042",
		NonceLength:    4,
		State:          model.AppAuthenticatorChallengeStatePending,
	}
}

func TestStorageAppAuthenticatorDeviceShouldRoundTrip(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	device := testAppDevice("device-1", "john")

	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, device))

	actual, err := provider.LoadAppAuthenticatorDeviceByDeviceID(ctx, "device-1")
	require.NoError(t, err)

	device.ID = actual.ID

	assert.True(t, device.CreatedAt.Equal(actual.CreatedAt))

	device.CreatedAt = actual.CreatedAt

	assert.Equal(t, device, *actual)
	assert.NotZero(t, actual.ID)

	byID, err := provider.LoadAppAuthenticatorDeviceByID(ctx, actual.ID)
	require.NoError(t, err)
	assert.Equal(t, *actual, *byID)

	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, testAppDevice("device-2", "john")))
	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, testAppDevice("device-3", "jane")))

	devices, err := provider.LoadAppAuthenticatorDevicesByUsername(ctx, "john")
	require.NoError(t, err)
	require.Len(t, devices, 2)
	assert.Equal(t, "device-1", devices[0].DeviceID)
	assert.Equal(t, "device-2", devices[1].DeviceID)

	_, err = provider.LoadAppAuthenticatorDeviceByDeviceID(ctx, "missing")
	assert.ErrorIs(t, err, sql.ErrNoRows)

	_, err = provider.LoadAppAuthenticatorDeviceByID(ctx, 9999)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestStorageAppAuthenticatorDeviceShouldRoundTripNullables(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	device := testAppDevice("device-1", "john")
	device.AttestationFormat = "none"
	device.AttestationIdentity = sql.NullString{}
	device.AttestationDetail = sql.NullString{}
	device.AttestationKeyID = nil
	device.PullIDPrevious = nil

	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, device))

	actual, err := provider.LoadAppAuthenticatorDeviceByDeviceID(ctx, "device-1")
	require.NoError(t, err)

	assert.False(t, actual.LastUsedAt.Valid)
	assert.False(t, actual.AttestationIdentity.Valid)
	assert.False(t, actual.AttestationDetail.Valid)
	assert.Empty(t, actual.AttestationKeyID)
	assert.Empty(t, actual.PullIDPrevious)
}

func TestStorageAppAuthenticatorDeviceShouldNotAllowDuplicateDeviceID(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, testAppDevice("device-1", "john")))
	assert.Error(t, provider.SaveAppAuthenticatorDevice(ctx, testAppDevice("device-1", "jane")))
}

func TestShouldNotRotatePullIDWhenCurrentChanged(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, testAppDevice("device-1", "john")))

	device, err := provider.LoadAppAuthenticatorDeviceByDeviceID(ctx, "device-1")
	require.NoError(t, err)

	updated, err := provider.UpdateAppAuthenticatorDevicePullIDs(ctx, device.ID, []byte{1, 1, 1}, []byte{2, 2, 2}, []byte{9, 9, 9})
	require.NoError(t, err)
	assert.False(t, updated)

	actual, err := provider.LoadAppAuthenticatorDeviceByID(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, device.PullIDCurrent, actual.PullIDCurrent)
	assert.Equal(t, device.PullIDPrevious, actual.PullIDPrevious)

	updated, err = provider.UpdateAppAuthenticatorDevicePullIDs(ctx, device.ID, []byte{9, 9, 9}, []byte{2, 2, 2}, []byte{9, 9, 9})
	require.NoError(t, err)
	assert.True(t, updated)

	actual, err = provider.LoadAppAuthenticatorDeviceByID(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, []byte{2, 2, 2}, actual.PullIDCurrent)
	assert.Equal(t, []byte{9, 9, 9}, actual.PullIDPrevious)
}

func TestStorageAppAuthenticatorDeviceShouldUpdateNameAndLastUsed(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, testAppDevice("device-1", "john")))

	device, err := provider.LoadAppAuthenticatorDeviceByDeviceID(ctx, "device-1")
	require.NoError(t, err)

	require.NoError(t, provider.UpdateAppAuthenticatorDeviceName(ctx, "jane", device.ID, "Stolen"))

	actual, err := provider.LoadAppAuthenticatorDeviceByID(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, "Jane's iPhone", actual.Name)

	require.NoError(t, provider.UpdateAppAuthenticatorDeviceName(ctx, "john", device.ID, "Work phone"))

	at := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)

	require.NoError(t, provider.UpdateAppAuthenticatorDeviceLastUsed(ctx, device.ID, at))

	actual, err = provider.LoadAppAuthenticatorDeviceByID(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, "Work phone", actual.Name)
	assert.True(t, actual.LastUsedAt.Valid)
	assert.True(t, at.Equal(actual.LastUsedAt.Time))
}

func TestStorageAppAuthenticatorRegistrationShouldRoundTripAndUpdate(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	registration := model.AppAuthenticatorRegistration{
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Minute * 5),
		Code:          []byte{1, 1},
		Username:      "john",
		PortalBinding: []byte{2, 2},
		BaseURL:       "https://auth.example.com",
		Number:        "0042",
		NonceLength:   4,
		State:         model.AppAuthenticatorRegistrationStateIssued,
	}

	require.NoError(t, provider.SaveAppAuthenticatorRegistration(ctx, registration))

	actual, err := provider.LoadAppAuthenticatorRegistrationByCode(ctx, []byte{1, 1})
	require.NoError(t, err)

	registration.ID = actual.ID

	assert.True(t, registration.CreatedAt.Equal(actual.CreatedAt))
	assert.True(t, registration.ExpiresAt.Equal(actual.ExpiresAt))

	registration.CreatedAt, registration.ExpiresAt = actual.CreatedAt, actual.ExpiresAt
	assert.Equal(t, registration, *actual)
	assert.Empty(t, actual.ConfirmationID)
	assert.False(t, actual.Challenge.Valid)

	actual.State = model.AppAuthenticatorRegistrationStateBegun
	actual.BegunAt = sql.NullTime{Time: now.Add(time.Second), Valid: true}
	actual.SessionExpiresAt = sql.NullTime{Time: now.Add(time.Minute), Valid: true}
	actual.ConfirmationID = []byte{3, 3}
	actual.Challenge = sql.NullString{String: "chal", Valid: true}
	actual.CodeChallenge = sql.NullString{String: "cc", Valid: true}
	actual.DevicePlatform = sql.NullString{String: "ios", Valid: true}
	actual.DeviceModel = sql.NullString{String: "iPhone", Valid: true}
	actual.DeviceName = sql.NullString{String: "Jane", Valid: true}
	actual.DeviceID = sql.NullString{String: "device-1", Valid: true}

	updated, err := provider.UpdateAppAuthenticatorRegistration(ctx, *actual, model.AppAuthenticatorRegistrationStateBegun)
	require.NoError(t, err)
	assert.False(t, updated)

	byCode, err := provider.LoadAppAuthenticatorRegistrationByCode(ctx, []byte{1, 1})
	require.NoError(t, err)
	assert.Equal(t, model.AppAuthenticatorRegistrationStateIssued, byCode.State)
	assert.False(t, byCode.Challenge.Valid)

	updated, err = provider.UpdateAppAuthenticatorRegistration(ctx, *actual, model.AppAuthenticatorRegistrationStateIssued)
	require.NoError(t, err)
	assert.True(t, updated)

	byConfirmation, err := provider.LoadAppAuthenticatorRegistrationByConfirmationID(ctx, []byte{3, 3})
	require.NoError(t, err)

	assert.Equal(t, model.AppAuthenticatorRegistrationStateBegun, byConfirmation.State)
	assert.Equal(t, "chal", byConfirmation.Challenge.String)
	assert.Equal(t, "device-1", byConfirmation.DeviceID.String)
	assert.True(t, now.Add(time.Second).Equal(byConfirmation.BegunAt.Time))
	assert.True(t, now.Add(time.Minute).Equal(byConfirmation.SessionExpiresAt.Time))

	_, err = provider.LoadAppAuthenticatorRegistrationByCode(ctx, []byte{9})
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestStorageAppAuthenticatorRegistrationShouldLoadNewestByPortalBindingAndCancel(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	for i, state := range []string{
		model.AppAuthenticatorRegistrationStateIssued,
		model.AppAuthenticatorRegistrationStateBegun,
		model.AppAuthenticatorRegistrationStateCompleted,
		model.AppAuthenticatorRegistrationStateIssued,
	} {
		binding := []byte{2, 2}

		if i == 3 {
			binding = []byte{5, 5}
		}

		require.NoError(t, provider.SaveAppAuthenticatorRegistration(ctx, model.AppAuthenticatorRegistration{
			CreatedAt:     now,
			ExpiresAt:     now.Add(time.Minute),
			Code:          []byte{byte(i)},
			Username:      "john",
			PortalBinding: binding,
			BaseURL:       "https://auth.example.com",
			Number:        "42",
			NonceLength:   2,
			State:         state,
		}))
	}

	newest, err := provider.LoadAppAuthenticatorRegistrationByPortalBinding(ctx, []byte{2, 2})
	require.NoError(t, err)
	assert.Equal(t, []byte{2}, newest.Code)

	require.NoError(t, provider.CancelAppAuthenticatorRegistrations(ctx, []byte{2, 2}))

	for code, expected := range map[byte]string{
		0: model.AppAuthenticatorRegistrationStateCancelled,
		1: model.AppAuthenticatorRegistrationStateCancelled,
		2: model.AppAuthenticatorRegistrationStateCompleted,
		3: model.AppAuthenticatorRegistrationStateIssued,
	} {
		r, err := provider.LoadAppAuthenticatorRegistrationByCode(ctx, []byte{code})
		require.NoError(t, err)
		assert.Equal(t, expected, r.State)
	}

	_, err = provider.LoadAppAuthenticatorRegistrationByPortalBinding(ctx, []byte{7})
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestStorageAppAuthenticatorRegistrationShouldEnforceNonceLength(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	now := time.Now().UTC()

	err := provider.SaveAppAuthenticatorRegistration(ctx, model.AppAuthenticatorRegistration{
		CreatedAt: now, ExpiresAt: now, Code: []byte{1}, Username: "john", PortalBinding: []byte{1}, BaseURL: "x", Number: "1", NonceLength: 1, State: "issued",
	})

	assert.Error(t, err)

	err = provider.SaveAppAuthenticatorRegistration(ctx, model.AppAuthenticatorRegistration{
		CreatedAt: now, ExpiresAt: now, Code: []byte{1}, Username: "john", PortalBinding: []byte{1}, BaseURL: "x", Number: "12345", NonceLength: 5, State: "issued",
	})

	assert.Error(t, err)
}

func TestStorageAppAuthenticatorChallengeShouldRoundTripAndUpdate(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	challenge := testAppChallenge("conf-1", "john", []byte{1, 2}, now.Add(time.Minute))
	challenge.PortalURL = sql.NullString{String: "https://auth.example.com", Valid: true}
	challenge.RemoteIP = sql.NullString{String: "192.0.2.1", Valid: true}

	require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, challenge))

	actual, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "conf-1")
	require.NoError(t, err)

	challenge.ID = actual.ID

	assert.True(t, challenge.CreatedAt.Equal(actual.CreatedAt))
	assert.True(t, challenge.ExpiresAt.Equal(actual.ExpiresAt))

	challenge.CreatedAt, challenge.ExpiresAt = actual.CreatedAt, actual.ExpiresAt
	assert.Equal(t, challenge, *actual)

	actual.State = model.AppAuthenticatorChallengeStateApproved
	actual.BoundDeviceID = sql.NullString{String: "device-1", Valid: true}
	actual.BoundCodeChallenge = sql.NullString{String: "cc", Valid: true}
	actual.DecidedAt = sql.NullTime{Time: now.Add(time.Second), Valid: true}
	actual.DecidedDeviceID = sql.NullString{String: "device-1", Valid: true}
	actual.ConsumedAt = sql.NullTime{Time: now.Add(2 * time.Second), Valid: true}
	actual.UserAgent = sql.NullString{String: "UA", Valid: true}
	actual.RequestedURL = sql.NullString{String: "https://app.example.com", Valid: true}

	updated, err := provider.UpdateAppAuthenticatorChallenge(ctx, *actual, model.AppAuthenticatorChallengeStateDenied)
	require.NoError(t, err)
	assert.False(t, updated)

	unchanged, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "conf-1")
	require.NoError(t, err)
	assert.Equal(t, model.AppAuthenticatorChallengeStatePending, unchanged.State)
	assert.False(t, unchanged.ConsumedAt.Valid)

	updated, err = provider.UpdateAppAuthenticatorChallenge(ctx, *actual, model.AppAuthenticatorChallengeStatePending)
	require.NoError(t, err)
	assert.True(t, updated)

	byBinding, err := provider.LoadAppAuthenticatorChallengeByPortalBinding(ctx, []byte{1, 2})
	require.NoError(t, err)

	assert.Equal(t, model.AppAuthenticatorChallengeStateApproved, byBinding.State)
	assert.Equal(t, "device-1", byBinding.BoundDeviceID.String)
	assert.Equal(t, "cc", byBinding.BoundCodeChallenge.String)
	assert.Equal(t, "UA", byBinding.UserAgent.String)
	assert.True(t, now.Add(time.Second).Equal(byBinding.DecidedAt.Time))
	assert.True(t, now.Add(2*time.Second).Equal(byBinding.ConsumedAt.Time))

	_, err = provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "missing")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestStorageAppAuthenticatorChallengeShouldLoadNewestByPortalBindingAndVoid(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, testAppChallenge("a", "john", []byte{1}, now.Add(time.Minute))))
	require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, testAppChallenge("b", "john", []byte{1}, now.Add(time.Minute))))

	approved := testAppChallenge("c", "john", []byte{1}, now.Add(time.Minute))
	approved.State = model.AppAuthenticatorChallengeStateApproved

	require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, approved))
	require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, testAppChallenge("d", "john", []byte{2}, now.Add(time.Minute))))

	newest, err := provider.LoadAppAuthenticatorChallengeByPortalBinding(ctx, []byte{1})
	require.NoError(t, err)
	assert.Equal(t, "c", newest.ConfirmationID)

	require.NoError(t, provider.VoidAppAuthenticatorChallengesByPortalBinding(ctx, []byte{1}))

	for id, expected := range map[string]string{
		"a": model.AppAuthenticatorChallengeStateVoided,
		"b": model.AppAuthenticatorChallengeStateVoided,
		"c": model.AppAuthenticatorChallengeStateApproved,
		"d": model.AppAuthenticatorChallengeStatePending,
	} {
		c, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, expected, c.State)
	}

	_, err = provider.LoadAppAuthenticatorChallengeByPortalBinding(ctx, []byte{9})
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestStorageAppAuthenticatorChallengePendingShouldReturnOldestUnexpired(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	_, err := provider.LoadAppAuthenticatorChallengePending(ctx, "john", now)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	expired := testAppChallenge("expired", "john", []byte{1}, now.Add(-time.Second))

	denied := testAppChallenge("denied", "john", []byte{1}, now.Add(time.Minute))
	denied.State = model.AppAuthenticatorChallengeStateDenied

	other := testAppChallenge("other", "jane", []byte{1}, now.Add(time.Minute))

	for _, c := range []model.AppAuthenticatorChallenge{
		expired,
		denied,
		other,
		testAppChallenge("first", "john", []byte{1}, now.Add(time.Minute)),
		testAppChallenge("second", "john", []byte{1}, now.Add(time.Hour)),
	} {
		require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, c))
	}

	actual, err := provider.LoadAppAuthenticatorChallengePending(ctx, "john", now)
	require.NoError(t, err)
	assert.Equal(t, "first", actual.ConfirmationID)

	actual, err = provider.LoadAppAuthenticatorChallengePending(ctx, "john", now.Add(time.Minute))
	require.NoError(t, err)
	assert.Equal(t, "first", actual.ConfirmationID)

	actual, err = provider.LoadAppAuthenticatorChallengePending(ctx, "john", now.Add(time.Minute+time.Second))
	require.NoError(t, err)
	assert.Equal(t, "second", actual.ConfirmationID)
}

func TestStorageAppAuthenticatorRequestNonceShouldInsertOnce(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	expires := time.Now().Add(time.Minute)

	inserted, err := provider.SaveAppAuthenticatorRequestNonce(ctx, "device-1", "nonce-aaaaaaaaaaaaaaaa", expires)
	require.NoError(t, err)
	assert.True(t, inserted)

	inserted, err = provider.SaveAppAuthenticatorRequestNonce(ctx, "device-1", "nonce-aaaaaaaaaaaaaaaa", expires)
	require.NoError(t, err)
	assert.False(t, inserted)

	inserted, err = provider.SaveAppAuthenticatorRequestNonce(ctx, "device-2", "nonce-aaaaaaaaaaaaaaaa", expires)
	require.NoError(t, err)
	assert.True(t, inserted)
}

func TestStorageAppAuthenticatorDeleteDeviceShouldClearBindingsAndNonces(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, testAppDevice("device-1", "john")))
	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, testAppDevice("device-2", "john")))

	for _, id := range []string{"device-1", "device-2"} {
		_, err := provider.SaveAppAuthenticatorRequestNonce(ctx, id, "nonce-aaaaaaaaaaaaaaaa", now.Add(time.Hour))
		require.NoError(t, err)
	}

	bound := testAppChallenge("bound", "john", []byte{1}, now.Add(time.Hour))
	bound.BoundDeviceID = sql.NullString{String: "device-1", Valid: true}
	bound.BoundCodeChallenge = sql.NullString{String: "cc", Valid: true}

	boundOther := testAppChallenge("other", "john", []byte{1}, now.Add(time.Hour))
	boundOther.BoundDeviceID = sql.NullString{String: "device-2", Valid: true}
	boundOther.BoundCodeChallenge = sql.NullString{String: "cc", Valid: true}

	approved := testAppChallenge("approved", "john", []byte{1}, now.Add(time.Hour))
	approved.State = model.AppAuthenticatorChallengeStateApproved
	approved.BoundDeviceID = sql.NullString{String: "device-1", Valid: true}

	for _, c := range []model.AppAuthenticatorChallenge{bound, boundOther, approved} {
		require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, c))
	}

	require.NoError(t, provider.DeleteAppAuthenticatorDevice(ctx, "device-1"))

	_, err := provider.LoadAppAuthenticatorDeviceByDeviceID(ctx, "device-1")
	assert.ErrorIs(t, err, sql.ErrNoRows)

	_, err = provider.LoadAppAuthenticatorDeviceByDeviceID(ctx, "device-2")
	assert.NoError(t, err)

	c, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "bound")
	require.NoError(t, err)
	assert.False(t, c.BoundDeviceID.Valid)
	assert.False(t, c.BoundCodeChallenge.Valid)

	c, err = provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "other")
	require.NoError(t, err)
	assert.Equal(t, "device-2", c.BoundDeviceID.String)

	c, err = provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "approved")
	require.NoError(t, err)
	assert.Equal(t, "device-1", c.BoundDeviceID.String)

	inserted, err := provider.SaveAppAuthenticatorRequestNonce(ctx, "device-1", "nonce-aaaaaaaaaaaaaaaa", now.Add(time.Hour))
	require.NoError(t, err)
	assert.True(t, inserted)

	inserted, err = provider.SaveAppAuthenticatorRequestNonce(ctx, "device-2", "nonce-aaaaaaaaaaaaaaaa", now.Add(time.Hour))
	require.NoError(t, err)
	assert.False(t, inserted)
}

func TestStorageAppAuthenticatorLoadUserInfoShouldReportHasMobileApp(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SavePreferred2FAMethod(ctx, "john", "totp"))

	info, err := provider.LoadUserInfo(ctx, "john")
	require.NoError(t, err)
	assert.False(t, info.HasMobileApp)

	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, testAppDevice("device-1", "john")))

	info, err = provider.LoadUserInfo(ctx, "john")
	require.NoError(t, err)
	assert.True(t, info.HasMobileApp)

	require.NoError(t, provider.SavePreferred2FAMethod(ctx, "jane", "totp"))

	info, err = provider.LoadUserInfo(ctx, "jane")
	require.NoError(t, err)
	assert.False(t, info.HasMobileApp)
}

func TestStorageAppAuthenticatorPruneShouldRemoveOnlyStaleRows(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

	_, err := provider.SaveAppAuthenticatorRequestNonce(ctx, "d", "old-aaaaaaaaaaaaaaaaaa", now.Add(-time.Second))
	require.NoError(t, err)

	_, err = provider.SaveAppAuthenticatorRequestNonce(ctx, "d", "new-aaaaaaaaaaaaaaaaaa", now.Add(time.Minute))
	require.NoError(t, err)

	reg := func(code byte, state string, expires time.Time) model.AppAuthenticatorRegistration {
		return model.AppAuthenticatorRegistration{
			CreatedAt: expires, ExpiresAt: expires, Code: []byte{code}, Username: "john", PortalBinding: []byte{1}, BaseURL: "x", Number: "42", NonceLength: 2, State: state,
		}
	}

	for _, r := range []model.AppAuthenticatorRegistration{
		reg(1, model.AppAuthenticatorRegistrationStateCompleted, now.Add(-48*time.Hour)),
		reg(2, model.AppAuthenticatorRegistrationStateCompleted, now.Add(-time.Hour)),
		reg(3, model.AppAuthenticatorRegistrationStateIssued, now.Add(-48*time.Hour)),
		reg(4, model.AppAuthenticatorRegistrationStateExpired, now.Add(-48*time.Hour)),
	} {
		require.NoError(t, provider.SaveAppAuthenticatorRegistration(ctx, r))
	}

	for _, c := range []model.AppAuthenticatorChallenge{
		testAppChallenge("old-denied", "john", []byte{1}, now.Add(-48*time.Hour)),
		testAppChallenge("old-pending", "john", []byte{1}, now.Add(-48*time.Hour)),
		testAppChallenge("recent", "john", []byte{1}, now.Add(-time.Hour)),
	} {
		if c.ConfirmationID == "old-denied" {
			c.State = model.AppAuthenticatorChallengeStateDenied
		}

		require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, c))
	}

	require.NoError(t, provider.PruneAppAuthenticator(ctx, now))

	inserted, err := provider.SaveAppAuthenticatorRequestNonce(ctx, "d", "old-aaaaaaaaaaaaaaaaaa", now.Add(time.Minute))
	require.NoError(t, err)
	assert.True(t, inserted)

	inserted, err = provider.SaveAppAuthenticatorRequestNonce(ctx, "d", "new-aaaaaaaaaaaaaaaaaa", now.Add(time.Minute))
	require.NoError(t, err)
	assert.False(t, inserted)

	for code, kept := range map[byte]bool{1: false, 2: true, 3: false, 4: false} {
		_, err = provider.LoadAppAuthenticatorRegistrationByCode(ctx, []byte{code})

		if kept {
			assert.NoError(t, err)
		} else {
			assert.ErrorIs(t, err, sql.ErrNoRows)
		}
	}

	for id, kept := range map[string]bool{"old-denied": false, "old-pending": false, "recent": true} {
		_, err = provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, id)

		if kept {
			assert.NoError(t, err)
		} else {
			assert.ErrorIs(t, err, sql.ErrNoRows)
		}
	}
}

func TestStorageAppAuthenticatorMigrationShouldGoDownAndUp(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SaveAppAuthenticatorDevice(ctx, testAppDevice("device-1", "john")))

	require.NoError(t, provider.SchemaMigrate(ctx, false, 33))

	version, err := provider.SchemaVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, 33, version)

	require.NoError(t, provider.SchemaMigrate(ctx, true, 34))

	version, err = provider.SchemaVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, 34, version)

	_, err = provider.LoadAppAuthenticatorDeviceByDeviceID(ctx, "device-1")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestStorageAppAuthenticatorChallengeShouldConsumeOnce(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	pending := testAppChallenge("pending", "john", []byte{1}, now.Add(time.Minute))
	approved := testAppChallenge("approved", "john", []byte{1}, now.Add(time.Minute))
	approved.State = model.AppAuthenticatorChallengeStateApproved

	require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, pending))
	require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, approved))

	p, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "pending")
	require.NoError(t, err)

	a, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "approved")
	require.NoError(t, err)

	consumed, err := provider.ConsumeAppAuthenticatorChallenge(ctx, p.ID, now, now)
	require.NoError(t, err)
	assert.False(t, consumed)

	consumed, err = provider.ConsumeAppAuthenticatorChallenge(ctx, a.ID, now, now.Add(time.Minute+time.Second))
	require.NoError(t, err)
	assert.False(t, consumed)

	consumed, err = provider.ConsumeAppAuthenticatorChallenge(ctx, a.ID, now, now.Add(time.Minute))
	require.NoError(t, err)
	assert.True(t, consumed)

	consumed, err = provider.ConsumeAppAuthenticatorChallenge(ctx, a.ID, now.Add(time.Second), now)
	require.NoError(t, err)
	assert.False(t, consumed)

	a, err = provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "approved")
	require.NoError(t, err)
	assert.True(t, now.Equal(a.ConsumedAt.Time))
}

func TestStorageAppAuthenticatorChallengeDecisionShouldRequireBinding(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	expires := time.Now().Add(time.Hour)

	require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, testAppChallenge("c-1", "john", []byte("binding"), expires)))

	loaded, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "c-1")
	require.NoError(t, err)

	decide := func(c model.AppAuthenticatorChallenge, state string) bool {
		c.State = state
		c.DecidedAt = sql.NullTime{Time: time.Now(), Valid: true}
		c.DecidedDeviceID = sql.NullString{String: "device-a", Valid: true}

		updated, err := provider.UpdateAppAuthenticatorChallengeDecision(ctx, c)
		require.NoError(t, err)

		return updated
	}

	stale := *loaded
	stale.BoundDeviceID = sql.NullString{String: "device-a", Valid: true}
	stale.BoundCodeChallenge = sql.NullString{String: "cc-a", Valid: true}

	assert.False(t, decide(stale, model.AppAuthenticatorChallengeStateApproved))

	bound := *loaded
	bound.BoundDeviceID = sql.NullString{String: "device-b", Valid: true}
	bound.BoundCodeChallenge = sql.NullString{String: "cc-b", Valid: true}

	updated, err := provider.UpdateAppAuthenticatorChallenge(ctx, bound, model.AppAuthenticatorChallengeStatePending)
	require.NoError(t, err)
	require.True(t, updated)

	assert.False(t, decide(*loaded, model.AppAuthenticatorChallengeStateApproved))
	assert.False(t, decide(stale, model.AppAuthenticatorChallengeStateApproved))

	other := bound
	other.BoundCodeChallenge = sql.NullString{String: "cc-other", Valid: true}

	assert.False(t, decide(other, model.AppAuthenticatorChallengeStateApproved))
	assert.True(t, decide(bound, model.AppAuthenticatorChallengeStateApproved))
	assert.False(t, decide(bound, model.AppAuthenticatorChallengeStateDenied))

	actual, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "c-1")
	require.NoError(t, err)
	assert.Equal(t, model.AppAuthenticatorChallengeStateApproved, actual.State)
	assert.Equal(t, "device-a", actual.DecidedDeviceID.String)
	assert.True(t, actual.DecidedAt.Valid)

	require.NoError(t, provider.SaveAppAuthenticatorChallenge(ctx, testAppChallenge("c-2", "john", []byte("binding"), expires)))

	unbound, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, "c-2")
	require.NoError(t, err)

	assert.True(t, decide(*unbound, model.AppAuthenticatorChallengeStateVoided))
}
