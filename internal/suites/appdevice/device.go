// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Package appdevice is a Go port of the Authelia app fake device. It holds software keys in memory and speaks the
// device side of the app authenticator protocol, for use in tests only.
package appdevice

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	schemeQR     = "authelia-app://"
	schemeQRHTTP = "authelia-app+http://"
	separator    = "::"
)

var (
	reQR         = regexp.MustCompile(`^((?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*|\[[0-9A-Fa-f:.]+\])(?::\d{1,5})?)((?:/[A-Za-z0-9._~-]+)*)/api/register/app/([A-Za-z0-9_-]+)$`)
	reID         = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)
	reChallenge  = regexp.MustCompile(`^[A-Za-z0-9_-]{1,512}$`)
	reDigits     = regexp.MustCompile(`^[0-9]+$`)
	reRemainder  = regexp.MustCompile(`^/[A-Za-z0-9._~/-]*(\?[A-Za-z0-9._~=&%-]*)?$`)
	reDefaultTLS = regexp.MustCompile(`:443$`)
)

var (
	// ErrInvalidQR is returned when the QR text is not an Authelia registration code.
	ErrInvalidQR = errors.New("invalid_qr")

	// ErrInvalidResponse is returned when the server response is not valid.
	ErrInvalidResponse = errors.New("invalid_response")

	// ErrBaseURLMismatch is returned when the server returns a URL outside the base URL.
	ErrBaseURLMismatch = errors.New("base_url_mismatch")

	// ErrRejected is returned when the device refuses to send a request, for example a number of the wrong length.
	ErrRejected = errors.New("rejected")

	// ErrStaleTimestamp is returned when the server rejects the timestamp twice.
	ErrStaleTimestamp = errors.New("stale_timestamp")
)

// HTTPError is a non-2xx response from the server.
type HTTPError struct {
	Status     int
	Code       string
	ServerTime int64
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("http %d: %s", e.Status, e.Code)
}

// BeginResult is the outcome of beginning a registration.
type BeginResult struct {
	BaseURL         string
	Challenge       string
	ConfirmationURL string
	ExpiresAt       int64
	Attestation     string
	NonceLength     int
}

// Challenge is a pending challenge pulled by a check.
type Challenge struct {
	Challenge      string
	ConfirmationID string
	ExpiresAt      int64
	NonceLength    int
	Details        map[string]any
	Verifier       string
}

// Info is the device information sent when a registration begins.
type Info struct {
	Platform string `json:"platform"`
	Model    string `json:"model"`
	Name     string `json:"name"`
}

// Option configures a Device.
type Option func(d *Device)

// WithHTTPClient sets the HTTP client used for every request.
func WithHTTPClient(client *http.Client) Option {
	return func(d *Device) {
		d.client = client
	}
}

// WithClock sets the clock of the device.
func WithClock(now func() time.Time) Option {
	return func(d *Device) {
		d.now = now
	}
}

// WithClockOffset shifts the clock of the device by the given offset.
func WithClockOffset(offset time.Duration) Option {
	return func(d *Device) {
		d.offset = offset
	}
}

// WithAllowHTTP allows the authelia-app+http QR form.
func WithAllowHTTP() Option {
	return func(d *Device) {
		d.allowHTTP = true
	}
}

// WithInfo sets the device information sent when a registration begins.
func WithInfo(info Info) Option {
	return func(d *Device) {
		d.info = info
	}
}

// WithAttestation sets the function which produces the attestation object sent at registration confirmation.
func WithAttestation(attest func(registrationString string) map[string]any) Option {
	return func(d *Device) {
		d.attest = attest
	}
}

// Device is an in-memory software device.
type Device struct {
	client    *http.Client
	now       func() time.Time
	offset    time.Duration
	allowHTTP bool
	info      Info
	attest    func(registrationString string) map[string]any

	pending  *BeginResult
	verifier string

	ApprovalKey  *ecdsa.PrivateKey
	TransportKey *ecdsa.PrivateKey

	BaseURL     string
	DeviceID    string
	PullID      string
	Username    string
	DisplayName string

	Last *Challenge

	// TamperRegistration, when set, is called with the registration string and the confirmation body before the body
	// is sent, allowing negative tests to alter it.
	TamperRegistration func(registrationString string, body map[string]any)
}

// New returns a new Device.
func New(opts ...Option) *Device {
	d := &Device{
		client: http.DefaultClient,
		now:    time.Now,
		info:   Info{Platform: "node", Model: "fake-device", Name: "fake-device"},
		attest: func(string) map[string]any {
			return map[string]any{"format": "none"}
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// SetClockOffset changes the clock offset of the device.
func (d *Device) SetClockOffset(offset time.Duration) {
	d.offset = offset
}

// ParseQR parses QR text into its base URL and registration URL.
func (d *Device) ParseQR(text string) (baseURL, registrationURL string, err error) {
	text = strings.TrimSpace(text)

	scheme := "https://"

	switch {
	case d.allowHTTP && strings.HasPrefix(text, schemeQRHTTP):
		scheme = "http://"
		text = strings.TrimPrefix(text, schemeQRHTTP)
	case strings.HasPrefix(text, schemeQR):
		text = strings.TrimPrefix(text, schemeQR)
	default:
		return "", "", ErrInvalidQR
	}

	m := reQR.FindStringSubmatch(text)
	if m == nil {
		return "", "", ErrInvalidQR
	}

	authority, subpath, random := m[1], m[2], m[3]

	for _, seg := range strings.Split(subpath, "/") {
		if seg == "." || seg == ".." {
			return "", "", ErrInvalidQR
		}
	}

	if i := strings.LastIndex(authority, ":"); i != -1 && !strings.HasSuffix(authority, "]") {
		if port, perr := strconv.Atoi(authority[i+1:]); perr != nil || port > 65535 {
			return "", "", ErrInvalidQR
		}
	}

	baseURL = scheme + reDefaultTLS.ReplaceAllString(strings.ToLower(authority), "") + subpath

	return baseURL, baseURL + "/api/register/app/" + random, nil
}

// Register runs both registration steps with a number known in advance.
func (d *Device) Register(qr, number string) error {
	if !isNumber(number, len(number)) {
		return ErrRejected
	}

	result, err := d.RegisterBegin(qr)
	if err != nil {
		return err
	}

	if len(number) != result.NonceLength {
		return ErrRejected
	}

	return d.RegisterConfirm(number)
}

// RegisterBegin begins a registration and creates the keys.
func (d *Device) RegisterBegin(qr string) (*BeginResult, error) {
	baseURL, registrationURL, err := d.ParseQR(qr)
	if err != nil {
		return nil, err
	}

	verifier, codeChallenge, err := newVerifier()
	if err != nil {
		return nil, err
	}

	body, err := d.postUnsigned(registrationURL, map[string]any{
		"code_challenge":        codeChallenge,
		"code_challenge_method": "S256",
		"device":                d.info,
	})
	if err != nil {
		return nil, err
	}

	result := &BeginResult{BaseURL: baseURL}

	if result.ConfirmationURL, err = str(body, "confirmation_url"); err != nil {
		return nil, err
	}

	if result.Attestation, err = str(body, "attestation"); err != nil {
		return nil, err
	}

	switch result.Attestation {
	case "required", "preferred", "none":
	default:
		return nil, ErrInvalidResponse
	}

	if !withinBase(baseURL, result.ConfirmationURL) {
		return nil, ErrBaseURLMismatch
	}

	if result.Challenge, err = match(body, "challenge", reChallenge); err != nil {
		return nil, err
	}

	if result.ExpiresAt, err = num(body, "expires_at"); err != nil {
		return nil, err
	}

	if result.NonceLength, err = nonceLength(body); err != nil {
		return nil, err
	}

	if d.ApprovalKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		return nil, err
	}

	if d.TransportKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		return nil, err
	}

	d.pending, d.verifier = result, verifier

	return result, nil
}

// SetRegistrationVerifier replaces the verifier of the begun registration, for negative tests.
func (d *Device) SetRegistrationVerifier(verifier string) {
	d.verifier = verifier
}

// RegisterConfirm confirms the begun registration with the number shown in the portal.
func (d *Device) RegisterConfirm(number string) error {
	if d.pending == nil {
		return ErrRejected
	}

	p := d.pending

	if !isNumber(number, p.NonceLength) {
		return ErrRejected
	}

	approvalSPKI, err := x509.MarshalPKIXPublicKey(&d.ApprovalKey.PublicKey)
	if err != nil {
		return err
	}

	transportSPKI, err := x509.MarshalPKIXPublicKey(&d.TransportKey.PublicKey)
	if err != nil {
		return err
	}

	message := strings.Join([]string{"authelia-app-v1:register", b64(p.BaseURL), p.Challenge, number, d.verifier, thumbprint(approvalSPKI), thumbprint(transportSPKI)}, separator)

	approvalSignature, err := sign(d.ApprovalKey, message)
	if err != nil {
		return err
	}

	transportSignature, err := sign(d.TransportKey, message)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"verifier":             d.verifier,
		"approval_public_key":  base64.RawURLEncoding.EncodeToString(approvalSPKI),
		"transport_public_key": base64.RawURLEncoding.EncodeToString(transportSPKI),
		"approval_signature":   approvalSignature,
		"transport_signature":  transportSignature,
		"attestation":          d.attest(message),
	}

	if d.TamperRegistration != nil {
		d.TamperRegistration(message, payload)
	}

	body, err := d.postUnsigned(p.ConfirmationURL, payload)
	if err != nil {
		return err
	}

	baseURL, err := str(body, "base_url")
	if err != nil {
		return err
	}

	if baseURL != p.BaseURL {
		return ErrBaseURLMismatch
	}

	if d.DeviceID, err = match(body, "device_id", reID); err != nil {
		return err
	}

	if d.PullID, err = match(body, "pull_id", reID); err != nil {
		return err
	}

	if d.Username, err = str(body, "username"); err != nil {
		return err
	}

	if d.DisplayName, err = str(body, "display_name"); err != nil {
		return err
	}

	d.BaseURL, d.pending, d.verifier = p.BaseURL, nil, ""

	return nil
}

// Check pulls a pending challenge, storing the next pull id before anything else.
func (d *Device) Check() (*Challenge, error) {
	verifier, codeChallenge, err := newVerifier()
	if err != nil {
		return nil, err
	}

	body, err := d.signedFetch(http.MethodPost, d.BaseURL+"/api/secondfactor/app/challenge/"+d.PullID, map[string]any{
		"code_challenge":        codeChallenge,
		"code_challenge_method": "S256",
	})
	if err != nil {
		return nil, err
	}

	if d.PullID, err = match(body, "pull_id", reID); err != nil {
		return nil, err
	}

	d.Last = nil

	raw, ok := body["challenge"]
	if !ok || raw == nil {
		return nil, nil
	}

	c, ok := raw.(map[string]any)
	if !ok {
		return nil, ErrInvalidResponse
	}

	challenge := &Challenge{Verifier: verifier}

	if challenge.Details, err = details(c); err != nil {
		return nil, err
	}

	if challenge.ExpiresAt, err = num(c, "expires_at"); err != nil {
		return nil, err
	}

	if challenge.Challenge, err = match(c, "challenge", reChallenge); err != nil {
		return nil, err
	}

	if challenge.ConfirmationID, err = match(c, "confirmation_id", reID); err != nil {
		return nil, err
	}

	if challenge.NonceLength, err = nonceLength(c); err != nil {
		return nil, err
	}

	d.Last = challenge

	return challenge, nil
}

// Approve approves the challenge from the last check with the number shown in the portal.
func (d *Device) Approve(number string) (string, error) {
	if d.Last == nil {
		return "", ErrRejected
	}

	c := d.Last

	result, err := d.Confirm(c, "approve", number, d.ApprovalKey)
	if err == nil || !errors.Is(err, ErrRejected) {
		d.Last = nil
	}

	return result, err
}

// Deny denies the challenge from the last check.
func (d *Device) Deny() (string, error) {
	if d.Last == nil {
		return "", ErrRejected
	}

	c := d.Last
	d.Last = nil

	return d.Confirm(c, "deny", "", nil)
}

// Confirm sends a decision for the given challenge. An approval is signed with the given key.
func (d *Device) Confirm(c *Challenge, decision, number string, key *ecdsa.PrivateKey) (string, error) {
	payload := map[string]any{"decision": decision, "verifier": c.Verifier}

	if decision == "approve" {
		if !isNumber(number, c.NonceLength) {
			return "", ErrRejected
		}

		signature, err := sign(key, strings.Join([]string{"authelia-app-v1:approve", b64(d.BaseURL), d.DeviceID, c.Challenge, number, c.Verifier}, separator))
		if err != nil {
			return "", err
		}

		payload["signature"] = signature
	}

	body, err := d.signedFetch(http.MethodPost, d.BaseURL+"/api/secondfactor/app/confirmation/"+c.ConfirmationID, payload)
	if err != nil {
		return "", err
	}

	result, _ := body["result"].(string)

	switch result {
	case "approved", "denied", "rejected":
		return result, nil
	default:
		return "", ErrInvalidResponse
	}
}

// Remove removes the device from the server.
func (d *Device) Remove() error {
	_, err := d.signedFetch(http.MethodDelete, d.BaseURL+"/api/secondfactor/app/device", nil)

	return err
}

// SignRequest builds a request signed with the transport key at the current device time. The returned request can
// be sent more than once by cloning it and resetting its body with GetBody.
func (d *Device) SignRequest(method, url string, body []byte) (*http.Request, error) {
	return d.signRequest(method, url, body, d.clock())
}

func (d *Device) signRequest(method, url string, body []byte, now time.Time) (*http.Request, error) {
	nonce := make([]byte, 16)

	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ts := strconv.FormatInt(now.Unix(), 10)
	requestNonce := base64.RawURLEncoding.EncodeToString(nonce)
	hash := sha256.Sum256(body)

	signed := strings.Join([]string{"authelia-app-v1:request", b64(d.BaseURL), d.DeviceID, strings.ToUpper(method), b64(url), ts, requestNonce, base64.RawURLEncoding.EncodeToString(hash[:])}, separator)

	sig, err := sign(d.TransportKey, signed)
	if err != nil {
		return nil, err
	}

	var reader io.Reader

	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf(`AutheliaApp device_id="%s", ts="%s", nonce="%s", sig="%s"`, d.DeviceID, ts, requestNonce, sig))

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// Do sends a request and decodes the JSON response body.
func (d *Device) Do(req *http.Request) (status int, body map[string]any, err error) {
	res, err := d.client.Do(req)
	if err != nil {
		return 0, nil, err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}

	body = map[string]any{}

	_ = json.Unmarshal(data, &body)

	return res.StatusCode, body, nil
}

func (d *Device) clock() time.Time {
	return d.now().Add(d.offset)
}

func (d *Device) signedFetch(method, url string, payload any) (map[string]any, error) {
	var data []byte

	if payload != nil {
		var err error

		if data, err = json.Marshal(payload); err != nil {
			return nil, err
		}
	}

	var offset time.Duration

	for attempt := 0; attempt < 2; attempt++ {
		req, err := d.signRequest(method, url, data, d.clock().Add(offset))
		if err != nil {
			return nil, err
		}

		status, body, err := d.Do(req)
		if err != nil {
			return nil, err
		}

		if status >= 200 && status < 300 {
			return body, nil
		}

		if status == http.StatusUnauthorized && body["error"] == "stale_timestamp" {
			if serverTime, ok := body["server_time"].(float64); ok && attempt == 0 {
				offset = time.Unix(int64(serverTime), 0).Sub(d.clock())

				continue
			}

			return nil, ErrStaleTimestamp
		}

		return nil, httpError(status, body)
	}

	return nil, ErrStaleTimestamp
}

func (d *Device) postUnsigned(url string, payload any) (map[string]any, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	status, body, err := d.Do(req)
	if err != nil {
		return nil, err
	}

	if status < 200 || status >= 300 {
		return nil, httpError(status, body)
	}

	return body, nil
}

func httpError(status int, body map[string]any) error {
	e := &HTTPError{Status: status, Code: fmt.Sprintf("HTTP %d", status)}

	if code, ok := body["error"].(string); ok {
		e.Code = code
	}

	if serverTime, ok := body["server_time"].(float64); ok {
		e.ServerTime = int64(serverTime)
	}

	return e
}

func newVerifier() (verifier, codeChallenge string, err error) {
	b := make([]byte, 32)

	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}

	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))

	return verifier, base64.RawURLEncoding.EncodeToString(h[:]), nil
}

func sign(key *ecdsa.PrivateKey, message string) (string, error) {
	h := sha256.Sum256([]byte(message))

	sig, err := ecdsa.SignASN1(rand.Reader, key, h[:])
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(sig), nil
}

func thumbprint(spki []byte) string {
	h := sha256.Sum256(spki)

	return base64.RawURLEncoding.EncodeToString(h[:])
}

func b64(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func isNumber(number string, length int) bool {
	return length >= 2 && length <= 4 && len(number) == length && reDigits.MatchString(number)
}

func withinBase(baseURL, candidate string) bool {
	base := strings.TrimRight(baseURL, "/")

	if candidate == base {
		return true
	}

	if !strings.HasPrefix(candidate, base+"/") {
		return false
	}

	remainder := candidate[len(base):]

	if !reRemainder.MatchString(remainder) {
		return false
	}

	path, _, _ := strings.Cut(remainder, "?")

	for _, seg := range strings.Split(path[1:], "/") {
		if seg == "" || seg == "." || seg == ".." {
			return false
		}
	}

	return true
}

func str(body map[string]any, key string) (string, error) {
	v, ok := body[key].(string)
	if !ok || v == "" {
		return "", fmt.Errorf("%w: %s", ErrInvalidResponse, key)
	}

	return v, nil
}

func match(body map[string]any, key string, re *regexp.Regexp) (string, error) {
	v, ok := body[key].(string)
	if !ok || !re.MatchString(v) {
		return "", fmt.Errorf("%w: %s", ErrInvalidResponse, key)
	}

	return v, nil
}

func num(body map[string]any, key string) (int64, error) {
	v, ok := body[key].(float64)
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrInvalidResponse, key)
	}

	return int64(v), nil
}

func details(c map[string]any) (map[string]any, error) {
	result := map[string]any{}

	raw, ok := c["details"]
	if !ok {
		return result, nil
	}

	values, ok := raw.(map[string]any)
	if !ok {
		return nil, ErrInvalidResponse
	}

	for _, k := range []string{"portal_url", "requested_url", "remote_ip", "user_agent"} {
		if v, ok := values[k].(string); ok {
			result[k] = v
		}
	}

	if v, ok := values["requested_at"].(float64); ok {
		result["requested_at"] = v
	}

	return result, nil
}

func nonceLength(body map[string]any) (int, error) {
	v, ok := body["nonce_length"].(float64)
	if !ok || v != float64(int(v)) || v < 2 || v > 4 {
		return 0, fmt.Errorf("%w: nonce_length", ErrInvalidResponse)
	}

	return int(v), nil
}
