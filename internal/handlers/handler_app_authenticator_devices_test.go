// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/templates"
)

func (p *appPortal) rowID(deviceID string) string {
	device, err := p.env.storage.LoadAppAuthenticatorDeviceByDeviceID(p.t.Context(), deviceID)
	require.NoError(p.t, err)

	return strconv.Itoa(device.ID)
}

func TestAppAuthenticatorDevicesShouldList(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")

	var devices []map[string]any

	appPortalOK(t, s.do(fasthttp.MethodGet, nil, handlers.AppAuthenticatorDevicesGET), &devices)
	assert.NotNil(t, devices)
	assert.Empty(t, devices)

	d := s.registerDevice()

	harry := p.login("harry")
	harry.registerDevice()

	appPortalOK(t, s.do(fasthttp.MethodGet, nil, handlers.AppAuthenticatorDevicesGET), &devices)
	require.Len(t, devices, 1)

	id, err := strconv.ParseFloat(p.rowID(d.DeviceID), 64)
	require.NoError(t, err)

	assert.Equal(t, id, devices[0]["id"])
	assert.Equal(t, "fake-device", devices[0]["name"])
	assert.Equal(t, "fake-device", devices[0]["model"])
	assert.Equal(t, "node", devices[0]["platform"])
	assert.Equal(t, "none", devices[0]["attestation_format"])
	assert.NotContains(t, devices[0], "device_id")
}

func TestAppAuthenticatorDevicesShouldRename(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	d := s.registerDevice()

	harry := p.login("harry")
	h := harry.registerDevice()

	id := p.rowID(d.DeviceID)

	appPortalOK(t, s.do(fasthttp.MethodPut, map[string]any{"name": "Work phone"}, handlers.AppAuthenticatorDevicePUT, "id", id), nil)

	device, err := p.env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), d.DeviceID)
	require.NoError(t, err)
	assert.Equal(t, "Work phone", device.Name)

	appPortalOK(t, s.do(fasthttp.MethodPut, map[string]any{"name": strings.Repeat("é", 64)}, handlers.AppAuthenticatorDevicePUT, "id", id), nil)

	for _, name := range []string{"", "   ", strings.Repeat("a", 65)} {
		appPortalKO(t, s.do(fasthttp.MethodPut, map[string]any{"name": name}, handlers.AppAuthenticatorDevicePUT, "id", id), fasthttp.StatusBadRequest)
	}

	appPortalKO(t, s.do(fasthttp.MethodPut, map[string]any{"name": 5}, handlers.AppAuthenticatorDevicePUT, "id", id), fasthttp.StatusBadRequest)
	appPortalKO(t, s.do(fasthttp.MethodPut, map[string]any{"name": "Phone"}, handlers.AppAuthenticatorDevicePUT, "id", "abc"), fasthttp.StatusBadRequest)
	appPortalKO(t, s.do(fasthttp.MethodPut, map[string]any{"name": "Phone"}, handlers.AppAuthenticatorDevicePUT, "id", "999999"), fasthttp.StatusNotFound)
	appPortalKO(t, s.do(fasthttp.MethodPut, map[string]any{"name": "Stolen"}, handlers.AppAuthenticatorDevicePUT, "id", p.rowID(h.DeviceID)), fasthttp.StatusNotFound)

	other, err := p.env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), h.DeviceID)
	require.NoError(t, err)
	assert.Equal(t, "fake-device", other.Name)
}

func TestAppAuthenticatorDevicesShouldDelete(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")
	d := s.registerDevice()

	harry := p.login("harry")
	h := harry.registerDevice()

	appPortalKO(t, s.do(fasthttp.MethodDelete, nil, handlers.AppAuthenticatorDeviceRemoveDELETE, "id", p.rowID(h.DeviceID)), fasthttp.StatusNotFound)

	_, err := p.env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), h.DeviceID)
	require.NoError(t, err)

	appPortalKO(t, s.do(fasthttp.MethodDelete, nil, handlers.AppAuthenticatorDeviceRemoveDELETE, "id", "abc"), fasthttp.StatusBadRequest)

	gomock.InOrder(
		p.mock.UserProviderMock.EXPECT().GetDetailsCached("john").Return(&authentication.UserDetails{Username: "john", DisplayName: "John Smith", Emails: []string{"john@example.com"}}, nil),
		p.mock.NotifierMock.EXPECT().Send(gomock.Any(), mail.Address{Name: "John Smith", Address: "john@example.com"}, "Second Factor Method Removed", gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, _ mail.Address, _ string, _ *templates.EmailTemplate, data any) error {
				values, ok := data.(templates.EmailEventValues)
				require.True(t, ok)
				assert.Equal(t, map[string]any{
					"Action":      "Second Factor Method Removed",
					"Category":    "Authelia app",
					"Description": "fake-device",
				}, values.Details)

				return nil
			}),
	)

	appPortalOK(t, s.do(fasthttp.MethodDelete, nil, handlers.AppAuthenticatorDeviceRemoveDELETE, "id", p.rowID(d.DeviceID)), nil)

	_, err = p.env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), d.DeviceID)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	_, err = d.Check()
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")

	_, err = h.Check()
	require.NoError(t, err)

	devices, err := p.env.storage.LoadAppAuthenticatorDevicesByUsername(t.Context(), "john")
	require.NoError(t, err)
	assert.Empty(t, devices)

	appPortalKO(t, s.do(fasthttp.MethodPost, appSignBody(), handlers.AppAuthenticatorPOST), fasthttp.StatusUnauthorized)

	var empty []model.AppAuthenticatorDeviceData

	appPortalOK(t, s.do(fasthttp.MethodGet, nil, handlers.AppAuthenticatorDevicesGET), &empty)
	assert.Empty(t, empty)
}

func TestAppAuthenticatorDevicesShouldRejectAnonymous(t *testing.T) {
	p := newAppPortal(t)
	s := &appPortalSession{p: p}

	appPortalKO(t, s.do(fasthttp.MethodGet, nil, handlers.AppAuthenticatorDevicesGET), fasthttp.StatusForbidden)
	appPortalKO(t, s.do(fasthttp.MethodPut, map[string]any{"name": "Phone"}, handlers.AppAuthenticatorDevicePUT, "id", "1"), fasthttp.StatusForbidden)
	appPortalKO(t, s.do(fasthttp.MethodDelete, nil, handlers.AppAuthenticatorDeviceRemoveDELETE, "id", "1"), fasthttp.StatusForbidden)
}
