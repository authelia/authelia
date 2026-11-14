// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/appauthenticator"
	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/server"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/suites/appdevice"
	"github.com/authelia/authelia/v4/internal/templates"
)

type appTestClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *appTestClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.now
}

func (c *appTestClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = c.now.Add(d)
}

func (c *appTestClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (c *appTestClock) AfterFunc(d time.Duration, f func()) *time.Timer {
	return time.AfterFunc(d, f)
}

func (c *appTestClock) GetJWTWithTimeFuncOption() jwt.ParserOption {
	return jwt.WithTimeFunc(c.Now)
}

type appTestLogStore struct {
	storage.Provider

	mu       sync.Mutex
	attempts []model.AuthenticationAttempt
}

func (s *appTestLogStore) AppendAuthenticationLog(ctx context.Context, attempt model.AuthenticationAttempt) error {
	s.mu.Lock()
	s.attempts = append(s.attempts, attempt)
	s.mu.Unlock()

	return s.Provider.AppendAuthenticationLog(ctx, attempt)
}

func (s *appTestLogStore) Attempts() []model.AuthenticationAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]model.AuthenticationAttempt(nil), s.attempts...)
}

type appTestHookStore struct {
	storage.Provider

	mu            sync.Mutex
	afterLoad     func(ctx context.Context, c *model.AppAuthenticatorChallenge)
	beforeRotate  func(ctx context.Context, id int)
	rotationCalls int
}

func (s *appTestHookStore) LoadAppAuthenticatorChallengeByConfirmationID(ctx context.Context, id string) (*model.AppAuthenticatorChallenge, error) {
	c, err := s.Provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, id)

	s.mu.Lock()
	hook := s.afterLoad
	s.afterLoad = nil
	s.mu.Unlock()

	if err == nil && hook != nil {
		hook(ctx, c)
	}

	return c, err
}

func (s *appTestHookStore) UpdateAppAuthenticatorDevicePullIDs(ctx context.Context, id int, expectedCurrent, current, previous []byte) (bool, error) {
	s.mu.Lock()
	hook := s.beforeRotate
	s.rotationCalls++
	s.mu.Unlock()

	if hook != nil {
		hook(ctx, id)
	}

	return s.Provider.UpdateAppAuthenticatorDevicePullIDs(ctx, id, expectedCurrent, current, previous)
}

func (s *appTestHookStore) setAfterLoad(hook func(ctx context.Context, c *model.AppAuthenticatorChallenge)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.afterLoad = hook
}

func (s *appTestHookStore) setBeforeRotate(hook func(ctx context.Context, id int)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.beforeRotate = hook
	s.rotationCalls = 0
}

func (s *appTestHookStore) rotations() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rotationCalls
}

type appTestUserProvider struct {
	authentication.UserProvider
}

func (appTestUserProvider) GetDetailsCached(username string) (*authentication.UserDetails, error) {
	return &authentication.UserDetails{Username: username, DisplayName: username, Emails: []string{username + "@example.com"}}, nil
}

type appTestNotification struct {
	Recipient string
	Subject   string
	Values    templates.EmailEventValues
}

type appTestNotifier struct {
	mu   sync.Mutex
	sent []appTestNotification
}

func (n *appTestNotifier) StartupCheck() error {
	return nil
}

func (n *appTestNotifier) Send(_ context.Context, recipient mail.Address, subject string, _ *templates.EmailTemplate, data any) error {
	values, _ := data.(templates.EmailEventValues)

	n.mu.Lock()
	n.sent = append(n.sent, appTestNotification{Recipient: recipient.Address, Subject: subject, Values: values})
	n.mu.Unlock()

	return nil
}

func (n *appTestNotifier) Sent() []appTestNotification {
	n.mu.Lock()
	defer n.mu.Unlock()

	return append([]appTestNotification(nil), n.sent...)
}

type appTestEnv struct {
	t        *testing.T
	notifier *appTestNotifier
	config   *schema.Configuration
	clock    *appTestClock
	storage  storage.Provider
	hooks    *appTestHookStore
	logs     *appTestLogStore
	addr     string
	origin   string
	base     string
	client   *http.Client
}

type appTestEnvOptions struct {
	path      string
	base      string
	mode      string
	nonce     int
	proxies   schema.ServerHeadersTrustedProxies
	rateLimit func(limits *schema.ServerEndpointRateLimits)
	client    func(addr string) *http.Client
}

type appTestEnvOption func(o *appTestEnvOptions)

func withAppMode(mode string) appTestEnvOption {
	return func(o *appTestEnvOptions) { o.mode = mode }
}

func withAppNonceLength(n int) appTestEnvOption {
	return func(o *appTestEnvOptions) { o.nonce = n }
}

func newAppTestEnv(t *testing.T, opts ...appTestEnvOption) *appTestEnv {
	t.Helper()

	o := &appTestEnvOptions{mode: schema.AppAuthenticatorAttestationModeNone, nonce: 2}

	for _, opt := range opts {
		opt(o)
	}

	ctx := t.Context()

	config := &schema.Configuration{
		Server: schema.Server{
			Address:   &schema.AddressTCP{Address: schema.NewAddressFromNetworkPathValues("tcp", "127.0.0.1", 0, o.path)},
			Buffers:   schema.DefaultServerConfiguration.Buffers,
			Endpoints: schema.DefaultServerConfiguration.Endpoints,
			Headers:   schema.ServerHeaders{TrustedProxies: o.proxies},
		},
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local:         &schema.StorageLocal{Path: filepath.Join(t.TempDir(), "db.sqlite3")},
		},
		Regulation:       schema.DefaultRegulationConfiguration,
		AppAuthenticator: schema.DefaultAppAuthenticatorConfiguration,
		DuoAPI:           schema.DuoAPI{Disable: true},
	}

	config.AppAuthenticator.Attestation.Mode = o.mode
	config.AppAuthenticator.NonceLength = o.nonce

	limits := &config.Server.Endpoints.RateLimits
	limits.AppAuthenticatorRegistration.Enable = false
	limits.AppAuthenticatorCheck.Enable = false
	limits.AppAuthenticatorConfirmation.Enable = false

	if o.rateLimit != nil {
		o.rateLimit(limits)
	}

	migrator, err := storage.NewSQLiteProvider(config)
	require.NoError(t, err)
	require.NoError(t, migrator.SchemaMigrate(ctx, true, storage.SchemaLatest))
	require.NoError(t, migrator.Close())

	store, err := storage.NewSQLiteProvider(config)
	require.NoError(t, err)

	t.Cleanup(func() { _ = store.Close() })

	clock := &appTestClock{now: time.Now().Truncate(time.Second)}
	logs := &appTestLogStore{Provider: store}

	tmpl, err := templates.New(templates.Config{})
	require.NoError(t, err)

	providers := middlewares.NewProvidersBasic()
	providers.Clock = clock
	hooks := &appTestHookStore{Provider: store}

	notifier := &appTestNotifier{}

	providers.StorageProvider = hooks
	providers.Templates = tmpl
	providers.UserProvider = appTestUserProvider{}
	providers.Notifier = notifier
	providers.Regulator = regulation.NewRegulator(config.Regulation, logs, clock)
	providers.AppAuthenticator = appauthenticator.NewVerifier(config.AppAuthenticator.Attestation, appauthenticator.WithClock(clock))

	srv, listener, _, _, err := server.New(config, providers)
	require.NoError(t, err)

	go func() { _ = srv.Serve(listener) }()

	t.Cleanup(func() { _ = srv.Shutdown() })

	env := &appTestEnv{
		t:        t,
		notifier: notifier,
		config:   config,
		clock:    clock,
		storage:  store,
		hooks:    hooks,
		logs:     logs,
		addr:     listener.Addr().String(),
		client:   &http.Client{Timeout: 10 * time.Second},
	}

	env.origin = "http://" + env.addr
	env.base = env.origin + o.path

	if o.base != "" {
		env.base = o.base
	}

	if o.client != nil {
		env.client = o.client(env.addr)
	}

	return env
}

func (e *appTestEnv) device(opts ...appdevice.Option) *appdevice.Device {
	return appdevice.New(append([]appdevice.Option{appdevice.WithAllowHTTP(), appdevice.WithClock(e.clock.Now), appdevice.WithHTTPClient(e.client)}, opts...)...)
}

func (e *appTestEnv) qr(random string) string {
	if strings.HasPrefix(e.base, "https://") {
		return "authelia-app://" + strings.TrimPrefix(e.base, "https://") + "/api/register/app/" + random
	}

	return "authelia-app+http://" + strings.TrimPrefix(e.base, "http://") + "/api/register/app/" + random
}

func appTestBinding(t *testing.T) []byte {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	require.NoError(t, err)

	return b
}

func (e *appTestEnv) registration(username string, nonceLength int) (qr, number string) {
	random, err := appauthenticator.NewID()
	require.NoError(e.t, err)

	number, err = appauthenticator.NewNumber(nonceLength)
	require.NoError(e.t, err)

	now := e.clock.Now()

	require.NoError(e.t, e.storage.SaveAppAuthenticatorRegistration(e.t.Context(), model.AppAuthenticatorRegistration{
		CreatedAt:     now,
		ExpiresAt:     now.Add(5 * time.Minute),
		Code:          appauthenticator.HashID(random),
		Username:      username,
		PortalBinding: appTestBinding(e.t),
		BaseURL:       e.base,
		Number:        number,
		NonceLength:   nonceLength,
		State:         model.AppAuthenticatorRegistrationStateIssued,
	}))

	return e.qr(random), number
}

func (e *appTestEnv) register(username string) *appdevice.Device {
	qr, number := e.registration(username, 2)

	d := e.device()
	require.NoError(e.t, d.Register(qr, number))

	return d
}

type appTestChallenge struct {
	ConfirmationID string
	Number         string
}

func (e *appTestEnv) challenge(username string, nonceLength int) appTestChallenge {
	confirmationID, err := appauthenticator.NewID()
	require.NoError(e.t, err)

	challenge, err := appauthenticator.NewID()
	require.NoError(e.t, err)

	number, err := appauthenticator.NewNumber(nonceLength)
	require.NoError(e.t, err)

	now := e.clock.Now()

	require.NoError(e.t, e.storage.SaveAppAuthenticatorChallenge(e.t.Context(), model.AppAuthenticatorChallenge{
		CreatedAt:      now,
		ExpiresAt:      now.Add(2 * time.Minute),
		ConfirmationID: confirmationID,
		Username:       username,
		PortalBinding:  appTestBinding(e.t),
		Challenge:      challenge,
		Number:         number,
		NonceLength:    nonceLength,
		State:          model.AppAuthenticatorChallengeStatePending,
		PortalURL:      sql.NullString{String: "https://auth.example.com", Valid: true},
		RequestedURL:   sql.NullString{String: "https://app.example.com/", Valid: true},
		RemoteIP:       sql.NullString{String: "203.0.113.7", Valid: true},
		UserAgent:      sql.NullString{String: "Firefox on macOS", Valid: true},
	}))

	return appTestChallenge{ConfirmationID: confirmationID, Number: number}
}

func (e *appTestEnv) row(c appTestChallenge) *model.AppAuthenticatorChallenge {
	row, err := e.storage.LoadAppAuthenticatorChallengeByConfirmationID(e.t.Context(), c.ConfirmationID)
	require.NoError(e.t, err)

	return row
}

func (e *appTestEnv) state(c appTestChallenge) string {
	return e.row(c).State
}

func (e *appTestEnv) send(req *http.Request) (int, map[string]any) {
	e.t.Helper()

	clone := req.Clone(e.t.Context())

	if req.GetBody != nil {
		body, err := req.GetBody()
		require.NoError(e.t, err)

		clone.Body = body
	}

	res, err := e.client.Do(clone)
	require.NoError(e.t, err)

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	require.NoError(e.t, err)

	body := map[string]any{}
	require.NoError(e.t, json.Unmarshal(data, &body), string(data))

	assert.Equal(e.t, "application/json", res.Header.Get("Content-Type"))

	return res.StatusCode, body
}

func (e *appTestEnv) post(url string, body []byte) (int, map[string]any) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	require.NoError(e.t, err)

	req.Header.Set("Content-Type", "application/json")

	return e.send(req)
}

func (e *appTestEnv) checkBody(t *testing.T) []byte {
	b, err := json.Marshal(map[string]any{"code_challenge": strings.Repeat("A", 43), "code_challenge_method": "S256"})
	require.NoError(t, err)

	return b
}

func requireHTTPError(t *testing.T, err error, status int, code string) {
	t.Helper()

	var e *appdevice.HTTPError

	require.True(t, errors.As(err, &e), "expected an http error but got %v", err)
	assert.Equal(t, status, e.Status)
	assert.Equal(t, code, e.Code)
	assert.NotZero(t, e.ServerTime)
}

func assertServerTime(t *testing.T, env *appTestEnv, body map[string]any) {
	t.Helper()

	assert.Equal(t, float64(env.clock.Now().Unix()), body["server_time"])
}

func wrongNumber(number string) string {
	n := 0

	for _, c := range number {
		n = n*10 + int(c-'0')
	}

	mod := 1

	for range number {
		mod *= 10
	}

	return fmt.Sprintf("%0*d", len(number), (n+1)%mod)
}

func TestAppAuthenticatorDeviceShouldRegisterAndApprove(t *testing.T) {
	env := newAppTestEnv(t)

	qr, number := env.registration("john", 2)
	assert.Regexp(t, `^[0-9]{2}$`, number)

	d := env.device()

	begin, err := d.RegisterBegin(qr)
	require.NoError(t, err)
	assert.Equal(t, env.base, begin.BaseURL)
	assert.Equal(t, "none", begin.Attestation)
	assert.Equal(t, 2, begin.NonceLength)
	assert.Equal(t, env.clock.Now().Add(5*time.Minute).Unix(), begin.ExpiresAt)
	assert.Regexp(t, `^`+env.base+`/api/register/app/confirm/[A-Za-z0-9_-]{43}$`, begin.ConfirmationURL)
	assert.Regexp(t, `^[A-Za-z0-9_-]{43}$`, begin.Challenge)

	require.NoError(t, d.RegisterConfirm(number))
	assert.Equal(t, env.base, d.BaseURL)
	assert.Equal(t, "john", d.Username)
	assert.Equal(t, "Authelia", d.DisplayName)

	device, err := env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), d.DeviceID)
	require.NoError(t, err)
	assert.Equal(t, "john", device.Username)
	assert.Equal(t, env.base, device.BaseURL)
	assert.Equal(t, "fake-device", device.Name)
	assert.Equal(t, "node", device.Platform)
	assert.Equal(t, appauthenticator.AttestationFormatNone, device.AttestationFormat)
	assert.Equal(t, appauthenticator.HashID(d.PullID), device.PullIDCurrent)
	assert.Nil(t, device.PullIDPrevious)

	c, err := d.Check()
	require.NoError(t, err)
	assert.Nil(t, c)

	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)
	require.NotNil(t, pulled)
	assert.Equal(t, ch.ConfirmationID, pulled.ConfirmationID)
	assert.Equal(t, 2, pulled.NonceLength)
	assert.Equal(t, env.clock.Now().Add(2*time.Minute).Unix(), pulled.ExpiresAt)
	assert.Equal(t, map[string]any{
		"portal_url":    "https://auth.example.com",
		"requested_url": "https://app.example.com/",
		"remote_ip":     "203.0.113.7",
		"user_agent":    "Firefox on macOS",
		"requested_at":  float64(env.clock.Now().Unix()),
	}, pulled.Details)
	assert.Equal(t, model.AppAuthenticatorChallengeStatePending, env.state(ch))

	result, err := d.Confirm(pulled, "approve", ch.Number, d.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "approved", result)

	row := env.row(ch)
	assert.Equal(t, model.AppAuthenticatorChallengeStateApproved, row.State)
	assert.Equal(t, d.DeviceID, row.DecidedDeviceID.String)
	assert.True(t, row.DecidedAt.Valid)

	result, err = d.Confirm(pulled, "approve", ch.Number, d.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateApproved, env.state(ch))

	result, err = d.Confirm(pulled, "deny", "", nil)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateApproved, env.state(ch))

	attempts := env.logs.Attempts()
	require.Len(t, attempts, 1)
	assert.True(t, attempts[0].Successful)
	assert.Equal(t, regulation.AuthTypeApp, attempts[0].Type)
	assert.Equal(t, "john", attempts[0].Username)

	require.NoError(t, d.Remove())

	_, err = env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), d.DeviceID)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	_, err = d.Check()
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")
}

func TestAppAuthenticatorDeviceShouldAuthenticateUnderBasePath(t *testing.T) {
	proxy := func(addr string) *http.Client {
		return &http.Client{
			Timeout: 10 * time.Second,
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				req = req.Clone(req.Context())
				req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
				req.Header.Set("X-Forwarded-Host", req.URL.Host)
				req.URL.Scheme = "http"
				req.URL.Host = addr
				req.Host = addr

				return http.DefaultTransport.RoundTrip(req)
			}),
		}
	}

	_, loopback, err := net.ParseCIDR("127.0.0.1/32")
	require.NoError(t, err)

	_, other, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	t.Run("ShouldApproveWithTrustedProxy", func(t *testing.T) {
		env := newAppTestEnv(t, func(o *appTestEnvOptions) {
			o.path = "/authelia"
			o.base = "https://auth.example.com/authelia"
			o.proxies = schema.ServerHeadersTrustedProxies{Enable: true, Networks: []*net.IPNet{loopback}}
			o.client = proxy
		})

		qr, number := env.registration("john", 2)
		assert.True(t, strings.HasPrefix(qr, "authelia-app://auth.example.com/authelia/api/register/app/"))

		d := env.device()

		begin, err := d.RegisterBegin(qr)
		require.NoError(t, err)
		assert.Equal(t, "https://auth.example.com/authelia", begin.BaseURL)
		assert.True(t, strings.HasPrefix(begin.ConfirmationURL, "https://auth.example.com/authelia/api/register/app/confirm/"))

		require.NoError(t, d.RegisterConfirm(number))
		assert.Equal(t, "https://auth.example.com/authelia", d.BaseURL)

		ch := env.challenge("john", 2)

		pulled, err := d.Check()
		require.NoError(t, err)
		require.NotNil(t, pulled)

		result, err := d.Approve(ch.Number)
		require.NoError(t, err)
		assert.Equal(t, "approved", result)

		require.NoError(t, d.Remove())

		_, err = d.Check()
		requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")
	})

	t.Run("ShouldRejectSignatureFromUntrustedProxy", func(t *testing.T) {
		env := newAppTestEnv(t, func(o *appTestEnvOptions) {
			o.path = "/authelia"
			o.base = "https://auth.example.com/authelia"
			o.proxies = schema.ServerHeadersTrustedProxies{Enable: true, Networks: []*net.IPNet{other}}
			o.client = proxy
		})

		d := env.register("john")

		_, err := d.Check()
		requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")
	})

	t.Run("ShouldRejectRequestOutsideBasePath", func(t *testing.T) {
		env := newAppTestEnv(t, func(o *appTestEnvOptions) {
			o.path = "/authelia"
		})

		qr, _ := env.registration("john", 2)

		status, body := env.post(strings.Replace(strings.Replace(qr, "authelia-app+http://", "http://", 1), "/authelia/", "/", 1), []byte(`{}`))
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "unknown", body["error"])
	})
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAppAuthenticatorRegistrationShouldVoidOnWrongNumber(t *testing.T) {
	env := newAppTestEnv(t)

	qr, number := env.registration("john", 2)

	d := env.device()

	_, err := d.RegisterBegin(qr)
	require.NoError(t, err)

	err = d.RegisterConfirm(wrongNumber(number))
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusConflict, "already_used")

	devices, err := env.storage.LoadAppAuthenticatorDevicesByUsername(t.Context(), "john")
	require.NoError(t, err)
	assert.Empty(t, devices)
}

func TestAppAuthenticatorRegistrationCodeShouldBeSingleUse(t *testing.T) {
	env := newAppTestEnv(t)

	qr, _ := env.registration("john", 2)

	_, err := env.device().RegisterBegin(qr)
	require.NoError(t, err)

	_, err = env.device().RegisterBegin(qr)
	requireHTTPError(t, err, http.StatusConflict, "already_used")
}

func TestAppAuthenticatorRegistrationCodeShouldExpire(t *testing.T) {
	env := newAppTestEnv(t)

	qr, _ := env.registration("john", 2)

	env.clock.Advance(301 * time.Second)

	_, err := env.device().RegisterBegin(qr)
	requireHTTPError(t, err, http.StatusGone, "expired")

	_, err = env.device().RegisterBegin(qr)
	requireHTTPError(t, err, http.StatusGone, "expired")
}

func TestAppAuthenticatorRegistrationSessionShouldExpire(t *testing.T) {
	env := newAppTestEnv(t)

	qr, number := env.registration("john", 2)

	d := env.device()

	_, err := d.RegisterBegin(qr)
	require.NoError(t, err)

	env.clock.Advance(301 * time.Second)

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusGone, "expired")

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusConflict, "already_used")
}

func TestAppAuthenticatorRegistrationShouldRejectNoneWhenRequired(t *testing.T) {
	env := newAppTestEnv(t, withAppMode(schema.AppAuthenticatorAttestationModeRequired))

	qr, number := env.registration("john", 2)

	d := env.device()

	begin, err := d.RegisterBegin(qr)
	require.NoError(t, err)
	assert.Equal(t, "required", begin.Attestation)

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusConflict, "already_used")

	_, err = env.device().RegisterBegin(qr)
	requireHTTPError(t, err, http.StatusConflict, "already_used")
}

func TestAppAuthenticatorRegistrationShouldAcceptNone(t *testing.T) {
	for _, mode := range []string{schema.AppAuthenticatorAttestationModeNone, schema.AppAuthenticatorAttestationModePreferred} {
		t.Run(mode, func(t *testing.T) {
			env := newAppTestEnv(t, withAppMode(mode))

			qr, number := env.registration("john", 2)

			d := env.device()

			begin, err := d.RegisterBegin(qr)
			require.NoError(t, err)
			assert.Equal(t, mode, begin.Attestation)

			require.NoError(t, d.RegisterConfirm(number))

			c, err := d.Check()
			require.NoError(t, err)
			assert.Nil(t, c)
		})
	}
}

func TestAppAuthenticatorRegistrationShouldVoidOnFailedAttestationWhenPreferred(t *testing.T) {
	env := newAppTestEnv(t, withAppMode(schema.AppAuthenticatorAttestationModePreferred))

	qr, number := env.registration("john", 2)

	attest := func(string) map[string]any {
		return map[string]any{"format": "android-key", "approval_certificate_chain": []string{"AA"}, "transport_certificate_chain": []string{"AA"}}
	}

	d := env.device(appdevice.WithAttestation(attest))

	_, err := d.RegisterBegin(qr)
	require.NoError(t, err)

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")

	appdevice.WithAttestation(func(string) map[string]any { return map[string]any{"format": "none"} })(d)

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusConflict, "already_used")
}

func TestAppAuthenticatorRegistrationShouldRejectMalformedRequests(t *testing.T) {
	env := newAppTestEnv(t)

	qr, number := env.registration("john", 2)

	d := env.device()

	_, registrationURL, err := d.ParseQR(qr)
	require.NoError(t, err)

	unknown, err := appauthenticator.NewID()
	require.NoError(t, err)

	testCases := []struct {
		name   string
		url    string
		body   string
		status int
		code   string
	}{
		{"ShouldRejectNonJSON", registrationURL, `not json`, http.StatusBadRequest, "invalid_request"},
		{"ShouldRejectNonObject", registrationURL, `[]`, http.StatusBadRequest, "invalid_request"},
		{"ShouldRejectMissingDevice", registrationURL, `{"code_challenge":"` + strings.Repeat("A", 43) + `","code_challenge_method":"S256"}`, http.StatusBadRequest, "invalid_request"},
		{"ShouldRejectShortCodeChallenge", registrationURL, `{"code_challenge":"AAAA","code_challenge_method":"S256","device":{}}`, http.StatusBadRequest, "invalid_request"},
		{"ShouldRejectPlainMethod", registrationURL, `{"code_challenge":"` + strings.Repeat("A", 43) + `","code_challenge_method":"plain","device":{}}`, http.StatusBadRequest, "invalid_request"},
		{"ShouldRejectUnknownCode", env.base + "/api/register/app/" + unknown, `{}`, http.StatusNotFound, "unknown"},
		{"ShouldRejectInvalidCode", env.base + "/api/register/app/a.b", `{}`, http.StatusNotFound, "unknown"},
		{"ShouldRejectUnknownSession", env.base + "/api/register/app/confirm/" + unknown, `{}`, http.StatusNotFound, "unknown"},
		{"ShouldRejectExtraSegment", env.base + "/api/register/app/confirm/" + unknown + "/x", `{}`, http.StatusNotFound, "unknown"},
		{"ShouldRejectQueryString", registrationURL + "?x=1", `{}`, http.StatusNotFound, "unknown"},
		{"ShouldRejectEncodedPath", strings.Replace(registrationURL, "/app/", "/%61pp/", 1), `{}`, http.StatusNotFound, "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := env.post(tc.url, []byte(tc.body))
			assert.Equal(t, tc.status, status)
			assert.Equal(t, tc.code, body["error"])
			assertServerTime(t, env, body)
		})
	}

	t.Run("ShouldRejectWrongMethod", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, registrationURL, nil)
		require.NoError(t, err)

		status, body := env.send(req)
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "unknown", body["error"])
	})

	begin, err := d.RegisterBegin(qr)
	require.NoError(t, err)

	for _, payload := range []string{
		`not json`,
		`{"verifier":"short","approval_public_key":"AA","transport_public_key":"AA","approval_signature":"AA","transport_signature":"AA","attestation":{"format":"none"}}`,
		`{"verifier":"` + strings.Repeat("A", 43) + `","approval_public_key":"","transport_public_key":"AA","approval_signature":"AA","transport_signature":"AA","attestation":{"format":"none"}}`,
		`{"verifier":"` + strings.Repeat("A", 43) + `","approval_public_key":"A+A","transport_public_key":"AA","approval_signature":"AA","transport_signature":"AA","attestation":{"format":"none"}}`,
		`{"verifier":"` + strings.Repeat("A", 43) + `","approval_public_key":"AA","transport_public_key":"AA","approval_signature":"AA","transport_signature":"AA","attestation":"none"}`,
		`{"verifier":"` + strings.Repeat("A", 43) + `","approval_public_key":"AA","transport_public_key":"AA","approval_signature":"AA","transport_signature":"AA","attestation":{"format":1}}`,
	} {
		status, body := env.post(begin.ConfirmationURL, []byte(payload))
		assert.Equal(t, http.StatusBadRequest, status, payload)
		assert.Equal(t, "invalid_request", body["error"], payload)
	}

	require.NoError(t, d.RegisterConfirm(number))

	status, body := env.post(registrationURL, []byte(`not json`))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_request", body["error"])
}

func TestAppAuthenticatorRegistrationShouldNotConsumeCodeOnMalformedBody(t *testing.T) {
	env := newAppTestEnv(t)

	qr, number := env.registration("john", 2)

	d := env.device()

	_, registrationURL, err := d.ParseQR(qr)
	require.NoError(t, err)

	status, body := env.post(registrationURL, []byte(`{"code_challenge":"AAAA","code_challenge_method":"S256","device":{}}`))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_request", body["error"])

	require.NoError(t, d.Register(qr, number))
}

func TestAppAuthenticatorRegistrationShouldRejectIdenticalKeys(t *testing.T) {
	env := newAppTestEnv(t)

	qr, number := env.registration("john", 2)

	d := env.device()

	_, err := d.RegisterBegin(qr)
	require.NoError(t, err)

	d.TransportKey = d.ApprovalKey

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")
}

func TestAppAuthenticatorChallengeShouldVoidOnWrongNumber(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)
	require.NotNil(t, pulled)

	result, err := d.Confirm(pulled, "approve", wrongNumber(ch.Number), d.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, env.state(ch))

	result, err = d.Confirm(pulled, "approve", ch.Number, d.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)

	c, err := d.Check()
	require.NoError(t, err)
	assert.Nil(t, c)

	attempts := env.logs.Attempts()
	require.Len(t, attempts, 1)
	assert.False(t, attempts[0].Successful)
	assert.Equal(t, regulation.AuthTypeApp, attempts[0].Type)
}

func TestAppAuthenticatorChallengeShouldVoidOnWrongLengthNumber(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)
	require.NotNil(t, pulled)

	forged := *pulled
	forged.NonceLength = 3

	result, err := d.Confirm(&forged, "approve", ch.Number+"0", d.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, env.state(ch))
}

func TestAppAuthenticatorChallengeShouldRejectWrongApprovalKey(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	result, err := d.Confirm(pulled, "approve", ch.Number, key)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, env.state(ch))
}

func TestAppAuthenticatorChallengeShouldOnlyConfirmLatestVerifier(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	first := env.challenge("john", 2)

	stale, err := d.Check()
	require.NoError(t, err)

	latest, err := d.Check()
	require.NoError(t, err)
	assert.Equal(t, stale.ConfirmationID, latest.ConfirmationID)

	result, err := d.Confirm(stale, "approve", first.Number, d.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, env.state(first))

	second := env.challenge("john", 2)

	_, err = d.Check()
	require.NoError(t, err)

	fresh, err := d.Check()
	require.NoError(t, err)

	result, err = d.Confirm(fresh, "approve", second.Number, d.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "approved", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateApproved, env.state(second))
}

func TestAppAuthenticatorChallengeShouldRejectExpired(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)

	env.clock.Advance(121 * time.Second)

	result, err := d.Confirm(pulled, "approve", ch.Number, d.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateExpired, env.state(ch))

	c, err := d.Check()
	require.NoError(t, err)
	assert.Nil(t, c)

	assert.Empty(t, env.logs.Attempts())
}

func TestAppAuthenticatorChallengeShouldAcceptAtExpiry(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)

	env.clock.Advance(120 * time.Second)

	result, err := d.Approve(ch.Number)
	require.NoError(t, err)
	assert.Equal(t, "approved", result)
	assert.Equal(t, ch.ConfirmationID, pulled.ConfirmationID)
}

func TestAppAuthenticatorChallengeShouldDeny(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)

	result, err := d.Confirm(pulled, "deny", "", nil)
	require.NoError(t, err)
	assert.Equal(t, "denied", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateDenied, env.state(ch))

	result, err = d.Confirm(pulled, "approve", ch.Number, d.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateDenied, env.state(ch))

	attempts := env.logs.Attempts()
	require.Len(t, attempts, 1)
	assert.False(t, attempts[0].Successful)
	assert.Equal(t, regulation.AuthTypeApp, attempts[0].Type)
	assert.Equal(t, "https://app.example.com/", attempts[0].RequestURI)
}

func TestAppAuthenticatorChallengeShouldIsolateUsers(t *testing.T) {
	env := newAppTestEnv(t)

	john := env.register("john")
	jane := env.register("jane")
	ch := env.challenge("john", 2)

	c, err := jane.Check()
	require.NoError(t, err)
	assert.Nil(t, c)

	pulled, err := john.Check()
	require.NoError(t, err)

	_, err = jane.Confirm(pulled, "deny", "", nil)
	requireHTTPError(t, err, http.StatusNotFound, "unknown")
	assert.Equal(t, model.AppAuthenticatorChallengeStatePending, env.state(ch))

	janes := env.challenge("jane", 2)

	c, err = jane.Check()
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, janes.ConfirmationID, c.ConfirmationID)

	pulled, err = john.Check()
	require.NoError(t, err)
	assert.Equal(t, ch.ConfirmationID, pulled.ConfirmationID)

	pullID := john.PullID
	jane.PullID = pullID

	_, err = jane.Check()
	requireHTTPError(t, err, http.StatusNotFound, "unknown")
}

func TestAppAuthenticatorChallengeShouldRebindToLastDevice(t *testing.T) {
	env := newAppTestEnv(t)

	a := env.register("john")
	b := env.register("john")
	ch := env.challenge("john", 2)

	pulledByA, err := a.Check()
	require.NoError(t, err)

	pulledByB, err := b.Check()
	require.NoError(t, err)
	assert.Equal(t, pulledByA.ConfirmationID, pulledByB.ConfirmationID)

	result, err := a.Confirm(pulledByA, "approve", ch.Number, a.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, env.state(ch))

	result, err = b.Confirm(pulledByB, "approve", ch.Number, b.ApprovalKey)
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
}

func TestAppAuthenticatorRemovedDeviceCannotCheck(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")

	require.NoError(t, d.Remove())

	_, err := d.Check()
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")

	err = d.Remove()
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")
}

func TestAppAuthenticatorRemovedDeviceCannotConfirm(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)

	row := env.row(ch)
	assert.Equal(t, d.DeviceID, row.BoundDeviceID.String)

	require.NoError(t, d.Remove())

	row = env.row(ch)
	assert.False(t, row.BoundDeviceID.Valid)
	assert.False(t, row.BoundCodeChallenge.Valid)
	assert.Equal(t, model.AppAuthenticatorChallengeStatePending, row.State)

	_, err = d.Confirm(pulled, "approve", ch.Number, d.ApprovalKey)
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")
	assert.Equal(t, model.AppAuthenticatorChallengeStatePending, env.state(ch))

	other := env.register("john")

	_, err = other.Check()
	require.NoError(t, err)

	result, err := other.Approve(ch.Number)
	require.NoError(t, err)
	assert.Equal(t, "approved", result)
}

func TestAppAuthenticatorPullIDShouldRotate(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	oldest := d.PullID

	_, err := d.Check()
	require.NoError(t, err)

	lost := d.PullID
	assert.NotEqual(t, oldest, lost)

	d.PullID = oldest

	_, err = d.Check()
	require.NoError(t, err)

	newest := d.PullID
	assert.NotEqual(t, oldest, newest)
	assert.NotEqual(t, lost, newest)

	d.PullID = lost

	_, err = d.Check()
	requireHTTPError(t, err, http.StatusNotFound, "unknown")

	d.PullID = newest

	_, err = d.Check()
	require.NoError(t, err)

	device, err := env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), d.DeviceID)
	require.NoError(t, err)
	assert.Equal(t, appauthenticator.HashID(d.PullID), device.PullIDCurrent)
	assert.Equal(t, appauthenticator.HashID(newest), device.PullIDPrevious)

	latest := d.PullID

	d.PullID = oldest

	_, err = d.Check()
	requireHTTPError(t, err, http.StatusNotFound, "unknown")

	d.PullID = newest

	_, err = d.Check()
	require.NoError(t, err)

	retried := d.PullID

	d.PullID = latest

	_, err = d.Check()
	requireHTTPError(t, err, http.StatusNotFound, "unknown")

	d.PullID = retried

	_, err = d.Check()
	require.NoError(t, err)
}

func TestAppAuthenticatorPullIDShouldRejectUnknown(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")

	d.PullID = "nope"

	_, err := d.Check()
	requireHTTPError(t, err, http.StatusNotFound, "unknown")
}

func TestAppAuthenticatorCheckShouldNotRotateOnMalformedBody(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	pullID := d.PullID

	for _, payload := range []string{`not json`, `{}`, `{"code_challenge":"AAAA","code_challenge_method":"S256"}`, `{"code_challenge":"` + strings.Repeat("A", 43) + `","code_challenge_method":"plain"}`} {
		req, err := d.SignRequest(http.MethodPost, d.BaseURL+"/api/secondfactor/app/challenge/"+pullID, []byte(payload))
		require.NoError(t, err)

		status, body := env.send(req)
		assert.Equal(t, http.StatusBadRequest, status, payload)
		assert.Equal(t, "invalid_request", body["error"], payload)
	}

	_, err := d.Check()
	require.NoError(t, err)

	device, err := env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), d.DeviceID)
	require.NoError(t, err)
	assert.Equal(t, appauthenticator.HashID(pullID), device.PullIDPrevious)
}

func TestAppAuthenticatorConfirmShouldRejectMalformedRequests(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)

	unknown, err := appauthenticator.NewID()
	require.NoError(t, err)

	testCases := []struct {
		name   string
		id     string
		body   string
		status int
		code   string
	}{
		{"ShouldRejectNonJSON", ch.ConfirmationID, `not json`, http.StatusBadRequest, "invalid_request"},
		{"ShouldRejectUnknownID", unknown, `{"decision":"deny"}`, http.StatusNotFound, "unknown"},
		{"ShouldRejectBadDecision", ch.ConfirmationID, `{"decision":"maybe","verifier":"` + pulled.Verifier + `"}`, http.StatusBadRequest, "invalid_request"},
		{"ShouldRejectMissingDecision", ch.ConfirmationID, `{"verifier":"` + pulled.Verifier + `"}`, http.StatusBadRequest, "invalid_request"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := d.SignRequest(http.MethodPost, d.BaseURL+"/api/secondfactor/app/confirmation/"+tc.id, []byte(tc.body))
			require.NoError(t, err)

			status, body := env.send(req)
			assert.Equal(t, tc.status, status)
			assert.Equal(t, tc.code, body["error"])
			assert.Equal(t, model.AppAuthenticatorChallengeStatePending, env.state(ch))
		})
	}

	req, err := d.SignRequest(http.MethodPost, d.BaseURL+"/api/secondfactor/app/confirmation/"+ch.ConfirmationID, []byte(`{"decision":"approve","verifier":"`+pulled.Verifier+`"}`))
	require.NoError(t, err)

	status, body := env.send(req)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "rejected", body["result"])
	assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, env.state(ch))
}

func TestAppAuthenticatorConfirmShouldVoidOnMalformedVerifier(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	ch := env.challenge("john", 2)

	_, err := d.Check()
	require.NoError(t, err)

	req, err := d.SignRequest(http.MethodPost, d.BaseURL+"/api/secondfactor/app/confirmation/"+ch.ConfirmationID, []byte(`{"decision":"deny","verifier":"short"}`))
	require.NoError(t, err)

	status, body := env.send(req)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "rejected", body["result"])
	assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, env.state(ch))
}

func TestAppAuthenticatorRequestShouldRejectStaleTimestamp(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")

	req, err := d.SignRequest(http.MethodPost, d.BaseURL+"/api/secondfactor/app/challenge/"+d.PullID, env.checkBody(t))
	require.NoError(t, err)

	env.clock.Advance(61 * time.Second)

	status, body := env.send(req)
	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Equal(t, "stale_timestamp", body["error"])
	assertServerTime(t, env, body)

	d.SetClockOffset(-61 * time.Second)

	req, err = d.SignRequest(http.MethodPost, d.BaseURL+"/api/secondfactor/app/challenge/"+d.PullID, env.checkBody(t))
	require.NoError(t, err)

	status, body = env.send(req)
	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Equal(t, "stale_timestamp", body["error"])
}

func TestAppAuthenticatorRequestShouldAcceptTimestampSixtySecondsOff(t *testing.T) {
	for _, offset := range []time.Duration{-60 * time.Second, 60 * time.Second} {
		t.Run(offset.String(), func(t *testing.T) {
			var count atomic.Int32

			env := newAppTestEnv(t, func(o *appTestEnvOptions) {
				o.client = func(string) *http.Client {
					return &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
						count.Add(1)

						return http.DefaultTransport.RoundTrip(req)
					})}
				}
			})

			d := env.register("john")

			d.SetClockOffset(offset)

			before := count.Load()

			c, err := d.Check()
			require.NoError(t, err)
			assert.Nil(t, c)
			assert.Equal(t, int32(1), count.Load()-before)
		})
	}
}

func TestAppAuthenticatorRequestShouldRecoverFromClockSkew(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")

	d.SetClockOffset(600 * time.Second)

	c, err := d.Check()
	require.NoError(t, err)
	assert.Nil(t, c)

	d.SetClockOffset(-1200 * time.Second)

	c, err = d.Check()
	require.NoError(t, err)
	assert.Nil(t, c)
}

func TestAppAuthenticatorRequestShouldRejectReplay(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")

	req, err := d.SignRequest(http.MethodPost, d.BaseURL+"/api/secondfactor/app/challenge/"+d.PullID, env.checkBody(t))
	require.NoError(t, err)

	status, body := env.send(req)
	require.Equal(t, http.StatusOK, status)
	assert.Nil(t, body["challenge"])
	assert.Contains(t, body, "challenge")
	assertServerTime(t, env, body)

	status, body = env.send(req)
	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Equal(t, "replay", body["error"])
	assertServerTime(t, env, body)
}

func TestAppAuthenticatorRequestShouldRejectTamperedBody(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")

	req, err := d.SignRequest(http.MethodPost, d.BaseURL+"/api/secondfactor/app/challenge/"+d.PullID, env.checkBody(t))
	require.NoError(t, err)

	tampered, err := http.NewRequest(http.MethodPost, req.URL.String(), strings.NewReader(`{"code_challenge":"`+strings.Repeat("B", 43)+`","code_challenge_method":"S256"}`))
	require.NoError(t, err)

	tampered.Header = req.Header.Clone()

	status, body := env.send(tampered)
	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Equal(t, "bad_signature", body["error"])

	status, _ = env.send(req)
	assert.Equal(t, http.StatusOK, status)
}

func TestAppAuthenticatorRequestShouldRejectBadAuthorization(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")
	url := d.BaseURL + "/api/secondfactor/app/challenge/" + d.PullID

	req, err := d.SignRequest(http.MethodPost, url, env.checkBody(t))
	require.NoError(t, err)

	header := req.Header.Get("Authorization")

	unknown, err := appauthenticator.NewID()
	require.NoError(t, err)

	testCases := []struct {
		name   string
		header string
	}{
		{"ShouldRejectMissing", ""},
		{"ShouldRejectScheme", strings.Replace(header, "AutheliaApp", "Bearer", 1)},
		{"ShouldRejectShortNonce", strings.Replace(header, `nonce="`, `nonce="A`, 1)},
		{"ShouldRejectSpacing", strings.Replace(header, `", ts=`, `",ts=`, 1)},
		{"ShouldRejectUnknownDevice", strings.Replace(header, d.DeviceID, unknown, 1)},
		{"ShouldRejectBadSignature", reSig.ReplaceAllString(header, `sig="${1}AA"`)},
		{"ShouldRejectUndecodableSignature", reSig.ReplaceAllString(header, `sig="A"`)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := req.Clone(t.Context())
			r.Header.Set("Authorization", tc.header)

			status, body := env.send(r)
			assert.Equal(t, http.StatusUnauthorized, status)
			assert.Equal(t, "bad_signature", body["error"])
			assertServerTime(t, env, body)
		})
	}

	status, _ := env.send(req)
	assert.Equal(t, http.StatusOK, status)
}

var reSig = regexp.MustCompile(`sig="([A-Za-z0-9_-]+)[A-Za-z0-9_-]{2}"`)

func TestAppAuthenticatorRequestShouldRejectOversizedBody(t *testing.T) {
	env := newAppTestEnv(t)

	qr, _ := env.registration("john", 2)

	d := env.device()

	_, registrationURL, err := d.ParseQR(qr)
	require.NoError(t, err)

	large := []byte(`{"pad":"` + strings.Repeat("x", 65*1024) + `"}`)

	status, body := env.post(registrationURL, large)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_request", body["error"])

	status, body = env.post(env.base+"/api/register/app/a.b", large)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_request", body["error"])

	status, body = env.post(env.base+"/api/register/app/confirm/a/b", large)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_request", body["error"])

	registered := env.register("john")

	req, err := registered.SignRequest(http.MethodPost, registered.BaseURL+"/api/secondfactor/app/challenge/"+registered.PullID, large)
	require.NoError(t, err)

	status, body = env.send(req)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_request", body["error"])
}

func TestAppAuthenticatorRequestShouldRejectQueryString(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")

	req, err := d.SignRequest(http.MethodPost, d.BaseURL+"/api/secondfactor/app/challenge/"+d.PullID+"?x=1", env.checkBody(t))
	require.NoError(t, err)

	status, body := env.send(req)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, "unknown", body["error"])

	req, err = d.SignRequest(http.MethodDelete, d.BaseURL+"/api/secondfactor/app/device?", nil)
	require.NoError(t, err)

	status, body = env.send(req)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, "unknown", body["error"])

	_, err = d.Check()
	require.NoError(t, err)
}

func TestAppAuthenticatorNonceLengthFour(t *testing.T) {
	env := newAppTestEnv(t, withAppNonceLength(4))

	qr, number := env.registration("john", 4)
	assert.Regexp(t, `^[0-9]{4}$`, number)

	d := env.device()

	begin, err := d.RegisterBegin(qr)
	require.NoError(t, err)
	assert.Equal(t, 4, begin.NonceLength)

	require.NoError(t, d.RegisterConfirm(number))

	ch := env.challenge("john", 4)
	assert.Regexp(t, `^[0-9]{4}$`, ch.Number)

	pulled, err := d.Check()
	require.NoError(t, err)
	assert.Equal(t, 4, pulled.NonceLength)

	result, err := d.Approve(ch.Number)
	require.NoError(t, err)
	assert.Equal(t, "approved", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateApproved, env.state(ch))
}

func TestAppAuthenticatorNonceLengthFourWrongNumber(t *testing.T) {
	env := newAppTestEnv(t, withAppNonceLength(4))

	qr, number := env.registration("john", 4)

	d := env.device()

	_, err := d.RegisterBegin(qr)
	require.NoError(t, err)

	err = d.RegisterConfirm(wrongNumber(number))
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")

	qr, number = env.registration("john", 4)

	d = env.device()
	require.NoError(t, d.Register(qr, number))

	ch := env.challenge("john", 4)

	_, err = d.Check()
	require.NoError(t, err)

	result, err := d.Approve(wrongNumber(ch.Number))
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)
	assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, env.state(ch))
}

func TestAppAuthenticatorChallengeShouldKeepStoredNonceLength(t *testing.T) {
	env := newAppTestEnv(t, withAppNonceLength(4))

	qr, number := env.registration("jane", 2)

	jane := env.device()

	begin, err := jane.RegisterBegin(qr)
	require.NoError(t, err)
	assert.Equal(t, 2, begin.NonceLength)

	require.NoError(t, jane.RegisterConfirm(number))
	assert.Equal(t, "jane", jane.Username)

	d := env.register("john")
	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)
	assert.Equal(t, 2, pulled.NonceLength)

	pulled, err = d.Check()
	require.NoError(t, err)
	assert.Equal(t, 2, pulled.NonceLength)

	result, err := d.Approve(ch.Number)
	require.NoError(t, err)
	assert.Equal(t, "approved", result)

	next := env.challenge("john", 3)

	pulled, err = d.Check()
	require.NoError(t, err)
	assert.Equal(t, 3, pulled.NonceLength)
	assert.Equal(t, next.ConfirmationID, pulled.ConfirmationID)
}

func TestAppAuthenticatorDeviceShouldUpdateLastUsed(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")

	lastUsed := func() time.Time {
		device, err := env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), d.DeviceID)
		require.NoError(t, err)
		require.True(t, device.LastUsedAt.Valid)

		return device.LastUsedAt.Time
	}

	first := env.clock.Now()

	_, err := d.Check()
	require.NoError(t, err)
	assert.True(t, first.Equal(lastUsed()))

	env.clock.Advance(30 * time.Second)

	_, err = d.Check()
	require.NoError(t, err)
	assert.True(t, first.Equal(lastUsed()))

	env.clock.Advance(31 * time.Second)

	_, err = d.Check()
	require.NoError(t, err)
	assert.True(t, env.clock.Now().Equal(lastUsed()))
}

func TestAppAuthenticatorDeviceShouldRateLimit(t *testing.T) {
	env := newAppTestEnv(t, func(o *appTestEnvOptions) {
		o.rateLimit = func(limits *schema.ServerEndpointRateLimits) {
			limits.AppAuthenticatorCheck = schema.ServerEndpointRateLimit{
				Enable:  true,
				Buckets: []schema.ServerEndpointRateLimitBucket{{Period: time.Minute, Requests: 1}},
			}
		}
	})

	d := env.register("john")

	_, err := d.Check()
	require.NoError(t, err)

	req, err := d.SignRequest(http.MethodPost, d.BaseURL+"/api/secondfactor/app/challenge/"+d.PullID, env.checkBody(t))
	require.NoError(t, err)

	status, body := env.send(req)
	assert.Equal(t, http.StatusTooManyRequests, status)
	assert.Equal(t, map[string]any{"error": "rate_limited", "server_time": float64(env.clock.Now().Unix())}, body)
}

func TestAppAuthenticatorDeviceShouldNotUseEnvelope(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")

	req, err := d.SignRequest(http.MethodDelete, d.BaseURL+"/api/secondfactor/app/device", nil)
	require.NoError(t, err)

	status, body := env.send(req)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, map[string]any{"server_time": float64(env.clock.Now().Unix())}, body)
}

func TestAppAuthenticatorDeviceShouldNotLogSecrets(t *testing.T) {
	logger := logging.Logger()
	level := logger.GetLevel()
	hook := logtest.NewLocal(logger)

	logger.SetLevel(logrus.TraceLevel)

	t.Cleanup(func() {
		logger.SetLevel(level)
		logger.ReplaceHooks(make(logrus.LevelHooks))
	})

	env := newAppTestEnv(t, func(o *appTestEnvOptions) {
		o.rateLimit = func(limits *schema.ServerEndpointRateLimits) {
			limits.AppAuthenticatorConfirmation = schema.ServerEndpointRateLimit{
				Enable:  true,
				Buckets: []schema.ServerEndpointRateLimitBucket{{Period: time.Minute, Requests: 2}},
			}
		}
	})

	d := env.register("john")
	secrets := make([]string, 0, 3)
	secrets = append(secrets, d.PullID)

	ch := env.challenge("john", 2)

	pulled, err := d.Check()
	require.NoError(t, err)

	secrets = append(secrets, d.PullID, pulled.Verifier)

	result, err := d.Approve(wrongNumber(ch.Number))
	require.NoError(t, err)
	assert.Equal(t, "rejected", result)

	d.PullID = "nope"

	_, err = d.Check()
	requireHTTPError(t, err, http.StatusNotFound, "unknown")

	_, err = d.Confirm(pulled, "deny", "", nil)
	require.NoError(t, err)

	_, err = d.Confirm(pulled, "deny", "", nil)
	requireHTTPError(t, err, http.StatusTooManyRequests, "rate_limited")

	entries := hook.AllEntries()
	require.NotEmpty(t, entries)

	for _, entry := range entries {
		line, err := entry.String()
		require.NoError(t, err)

		for _, secret := range secrets {
			assert.NotContains(t, line, secret)
		}

		assert.NotContains(t, line, ch.ConfirmationID)
	}
}

func TestAppAuthenticatorRegistrationShouldTruncateDeviceInfo(t *testing.T) {
	env := newAppTestEnv(t)

	qr, number := env.registration("john", 2)

	d := env.device(appdevice.WithInfo(appdevice.Info{Platform: strings.Repeat("p", 40), Model: strings.Repeat("m", 100), Name: strings.Repeat("é", 100)}))

	require.NoError(t, d.Register(qr, number))

	device, err := env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), d.DeviceID)
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat("p", 16), device.Platform)
	assert.Equal(t, strings.Repeat("m", 64), device.Model)
	assert.Equal(t, strings.Repeat("é", 64), device.Name)
}

func TestAppAuthenticatorRegistrationShouldVoidOnWrongVerifier(t *testing.T) {
	env := newAppTestEnv(t)

	qr, number := env.registration("john", 2)

	d := env.device()

	_, err := d.RegisterBegin(qr)
	require.NoError(t, err)

	wrong, err := appauthenticator.NewID()
	require.NoError(t, err)

	d.SetRegistrationVerifier(wrong)

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusConflict, "already_used")
}

func TestAppAuthenticatorRegistrationShouldVoidOnBadTransportSignature(t *testing.T) {
	env := newAppTestEnv(t)

	qr, number := env.registration("john", 2)

	d := env.device()

	_, err := d.RegisterBegin(qr)
	require.NoError(t, err)

	other, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	d.TamperRegistration = func(message string, body map[string]any) {
		h := sha256.Sum256([]byte(message))

		sig, err := ecdsa.SignASN1(rand.Reader, other, h[:])
		require.NoError(t, err)

		body["transport_signature"] = base64.RawURLEncoding.EncodeToString(sig)
	}

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")

	d.TamperRegistration = nil

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusConflict, "already_used")
}

func TestAppAuthenticatorRegistrationCodeShouldBeUsedAfterSessionExpiry(t *testing.T) {
	env := newAppTestEnv(t)

	qr, number := env.registration("john", 2)

	d := env.device()

	_, err := d.RegisterBegin(qr)
	require.NoError(t, err)

	env.clock.Advance(301 * time.Second)

	err = d.RegisterConfirm(number)
	requireHTTPError(t, err, http.StatusGone, "expired")

	_, err = env.device().RegisterBegin(qr)
	requireHTTPError(t, err, http.StatusConflict, "already_used")

	env.clock.Advance(time.Hour)

	_, err = env.device().RegisterBegin(qr)
	requireHTTPError(t, err, http.StatusConflict, "already_used")
}

func TestAppAuthenticatorChallengeShouldVoidWhenBindingChangesBeforeSettle(t *testing.T) {
	t.Run("ShouldVoidOnRebind", func(t *testing.T) {
		env := newAppTestEnv(t)

		a := env.register("john")
		b := env.register("john")
		ch := env.challenge("john", 2)

		pulled, err := a.Check()
		require.NoError(t, err)

		env.hooks.setAfterLoad(func(_ context.Context, c *model.AppAuthenticatorChallenge) {
			rebound := *c
			rebound.BoundDeviceID = sql.NullString{String: b.DeviceID, Valid: true}
			rebound.BoundCodeChallenge = sql.NullString{String: strings.Repeat("B", 43), Valid: true}

			updated, err := env.storage.UpdateAppAuthenticatorChallenge(t.Context(), rebound, model.AppAuthenticatorChallengeStatePending)
			assert.NoError(t, err)
			assert.True(t, updated)
		})

		result, err := a.Confirm(pulled, "approve", ch.Number, a.ApprovalKey)
		require.NoError(t, err)
		assert.Equal(t, "rejected", result)

		row := env.row(ch)
		assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, row.State)
		assert.Equal(t, a.DeviceID, row.DecidedDeviceID.String)

		attempts := env.logs.Attempts()
		require.Len(t, attempts, 1)
		assert.False(t, attempts[0].Successful)
	})

	t.Run("ShouldVoidOnRemoval", func(t *testing.T) {
		env := newAppTestEnv(t)

		d := env.register("john")
		ch := env.challenge("john", 2)

		pulled, err := d.Check()
		require.NoError(t, err)

		env.hooks.setAfterLoad(func(_ context.Context, _ *model.AppAuthenticatorChallenge) {
			assert.NoError(t, env.storage.DeleteAppAuthenticatorDevice(t.Context(), d.DeviceID))
		})

		result, err := d.Confirm(pulled, "approve", ch.Number, d.ApprovalKey)
		require.NoError(t, err)
		assert.Equal(t, "rejected", result)
		assert.Equal(t, model.AppAuthenticatorChallengeStateVoided, env.state(ch))
	})
}

func TestAppAuthenticatorPullIDShouldRestartLostRotation(t *testing.T) {
	compete := func(env *appTestEnv, keepPrevious bool) func(context.Context, int) {
		return func(ctx context.Context, id int) {
			device, err := env.storage.LoadAppAuthenticatorDeviceByID(ctx, id)
			if !assert.NoError(env.t, err) {
				return
			}

			competitor, err := appauthenticator.NewID()
			if !assert.NoError(env.t, err) {
				return
			}

			previous := device.PullIDCurrent
			if keepPrevious {
				previous = device.PullIDPrevious
			}

			updated, err := env.storage.UpdateAppAuthenticatorDevicePullIDs(ctx, id, device.PullIDCurrent, appauthenticator.HashID(competitor), previous)
			assert.NoError(env.t, err)
			assert.True(env.t, updated)
		}
	}

	t.Run("ShouldSucceedAfterRestart", func(t *testing.T) {
		env := newAppTestEnv(t)

		d := env.register("john")
		original := d.PullID

		race := compete(env, false)
		once := sync.Once{}

		env.hooks.setBeforeRotate(func(ctx context.Context, id int) {
			once.Do(func() { race(ctx, id) })
		})

		_, err := d.Check()
		require.NoError(t, err)
		assert.Equal(t, 2, env.hooks.rotations())

		device, err := env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), d.DeviceID)
		require.NoError(t, err)
		assert.Equal(t, appauthenticator.HashID(d.PullID), device.PullIDCurrent)
		assert.Equal(t, appauthenticator.HashID(original), device.PullIDPrevious)
	})

	t.Run("ShouldAnswerUnknownWhenEveryAttemptIsLost", func(t *testing.T) {
		env := newAppTestEnv(t)

		d := env.register("john")

		_, err := d.Check()
		require.NoError(t, err)

		previous := d.PullID

		_, err = d.Check()
		require.NoError(t, err)

		d.PullID = previous

		env.hooks.setBeforeRotate(compete(env, true))

		_, err = d.Check()
		requireHTTPError(t, err, http.StatusNotFound, "unknown")
		assert.Equal(t, 3, env.hooks.rotations())
	})
}

func TestAppAuthenticatorRequestShouldIgnoreForwardedURIForOrigin(t *testing.T) {
	env := newAppTestEnv(t, func(o *appTestEnvOptions) {
		o.client = func(string) *http.Client {
			return &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				req = req.Clone(req.Context())
				req.Header.Set("X-Forwarded-URI", "@other.example/")

				return http.DefaultTransport.RoundTrip(req)
			})}
		}
	})

	d := env.register("john")

	c, err := d.Check()
	require.NoError(t, err)
	assert.Nil(t, c)
}

func TestAppAuthenticatorDeviceShouldAnswerUnknownForOtherMethods(t *testing.T) {
	env := newAppTestEnv(t)

	qr, _ := env.registration("john", 2)

	d := env.device()

	begin, err := d.RegisterBegin(qr)
	require.NoError(t, err)

	_, registrationURL, err := d.ParseQR(qr)
	require.NoError(t, err)

	urls := []string{
		registrationURL,
		begin.ConfirmationURL,
		env.base + "/api/secondfactor/app/challenge/" + strings.Repeat("A", 43),
		env.base + "/api/secondfactor/app/confirmation/" + strings.Repeat("A", 43),
		env.base + "/api/secondfactor/app/device",
		env.base + "/api/secondfactor/app/device/extra",
	}

	for _, method := range []string{http.MethodOptions, http.MethodPut, http.MethodPatch, http.MethodGet} {
		for _, url := range urls {
			t.Run(method+url[len(env.base):], func(t *testing.T) {
				req, err := http.NewRequest(method, url, nil)
				require.NoError(t, err)

				status, body := env.send(req)
				assert.Equal(t, http.StatusNotFound, status)
				assert.Equal(t, "unknown", body["error"])
				assertServerTime(t, env, body)
			})
		}
	}
}

func TestAppAuthenticatorDeviceShouldNotifyOnRegistrationAndRemoval(t *testing.T) {
	env := newAppTestEnv(t)

	d := env.register("john")

	sent := env.notifier.Sent()
	require.Len(t, sent, 1)
	assert.Equal(t, "john@example.com", sent[0].Recipient)
	assert.Equal(t, "Second Factor Method Added", sent[0].Subject)
	assert.Equal(t, "was added to your account.", sent[0].Values.BodySuffix)
	assert.Equal(t, map[string]any{
		"Action":      "Second Factor Method Added",
		"Category":    "Authelia app",
		"Description": "fake-device",
	}, sent[0].Values.Details)

	require.NoError(t, d.Remove())

	sent = env.notifier.Sent()
	require.Len(t, sent, 2)
	assert.Equal(t, "john@example.com", sent[1].Recipient)
	assert.Equal(t, "Second Factor Method Removed", sent[1].Subject)
	assert.Equal(t, "was removed from your account.", sent[1].Values.BodySuffix)
	assert.Equal(t, map[string]any{
		"Action":      "Second Factor Method Removed",
		"Category":    "Authelia app",
		"Description": "fake-device",
	}, sent[1].Values.Details)

	err := d.Remove()
	requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")

	assert.Len(t, env.notifier.Sent(), 2)
}
