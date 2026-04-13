package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// AutheliaClient handles HTTP communication with the Authelia server.
type AutheliaClient struct {
	client        *http.Client
	baseURL       string
	cookieName    string
	sessionCookie string
	debug         bool
}

// UserInfoResponse represents the response from the user info endpoint.
type UserInfoResponse struct {
	DisplayName string `json:"display_name"`
	Method      string `json:"method"`
	HasTOTP     bool   `json:"has_totp"`
	HasWebAuthn bool   `json:"has_webauthn"`
	HasDuo      bool   `json:"has_duo"`
}

type apiResponse struct {
	Status string `json:"status"`
}

// NewAutheliaClient creates a new client for the Authelia API.
func NewAutheliaClient(cfg *Config) (*AutheliaClient, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	if cfg.CACert != "" {
		caCert, err := os.ReadFile(cfg.CACert) //nolint:gosec // Path comes from trusted PAM module configuration.
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		pool := x509.NewCertPool()

		if !pool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to parse CA certificate")
		}

		transport.TLSClientConfig.RootCAs = pool
	}

	return &AutheliaClient{
		client: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		baseURL:    strings.TrimRight(cfg.URL.String(), "/"),
		cookieName: cfg.CookieName,
		debug:      cfg.Debug,
	}, nil
}

// FirstFactor performs first-factor authentication and stores the session cookie.
func (c *AutheliaClient) FirstFactor(username, password string) error {
	body := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := c.postJSON("/api/firstfactor", body)
	if err != nil {
		return fmt.Errorf("first factor request failed: %w", err)
	}

	defer resp.Body.Close()

	c.extractSessionCookie(resp)

	return c.checkResponse(resp, "first factor authentication failed")
}

// UserInfo retrieves the user's 2FA method information.
func (c *AutheliaClient) UserInfo() (*UserInfoResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/user/info", nil) //nolint:gosec // URL from trusted PAM config.
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.client.Do(req) //nolint:gosec // URL from trusted PAM config.
	if err != nil {
		return nil, fmt.Errorf("user info request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request returned status %d", resp.StatusCode)
	}

	c.extractSessionCookie(resp)

	var info UserInfoResponse

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<14))
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	c.debugf("user info response: %s", string(data))

	var envelope struct {
		Status string           `json:"status"`
		Data   UserInfoResponse `json:"data"`
	}

	if err = json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	info = envelope.Data

	return &info, nil
}

// SecondFactorTOTP performs TOTP second-factor authentication.
func (c *AutheliaClient) SecondFactorTOTP(token string) error {
	body := map[string]string{
		"token": token,
	}

	resp, err := c.postJSON("/api/secondfactor/totp", body)
	if err != nil {
		return fmt.Errorf("TOTP request failed: %w", err)
	}

	defer resp.Body.Close()

	c.extractSessionCookie(resp)

	return c.checkResponse(resp, "TOTP authentication failed")
}

// SecondFactorDuoPush performs Duo push second-factor authentication; it blocks
// until the user approves, denies, or the request times out server-side.
func (c *AutheliaClient) SecondFactorDuoPush() error {
	body := map[string]string{}

	resp, err := c.postJSON("/api/secondfactor/duo", body)
	if err != nil {
		return fmt.Errorf("duo push request failed: %w", err)
	}

	defer resp.Body.Close()

	c.extractSessionCookie(resp)

	return c.checkResponse(resp, "duo push authentication failed")
}

//nolint:gosec // URL constructed from trusted PAM configuration, not user input.
func (c *AutheliaClient) postJSON(path string, body any) (*http.Response, error) {
	c.debugf("POST %s%s", c.baseURL, path)

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(payload)) //nolint:gosec // URL from trusted PAM config.
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	return c.client.Do(req)
}

func (c *AutheliaClient) setHeaders(req *http.Request) {
	if c.sessionCookie != "" {
		req.AddCookie(&http.Cookie{
			Name:  c.cookieName,
			Value: c.sessionCookie,
		})
	}
}

func (c *AutheliaClient) extractSessionCookie(resp *http.Response) {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == c.cookieName {
			c.sessionCookie = cookie.Value

			return
		}
	}
}

func (c *AutheliaClient) debugf(format string, args ...any) {
	if c.debug {
		fmt.Fprintf(os.Stderr, "authelia-pam: "+format+"\n", args...)
	}
}

func (c *AutheliaClient) checkResponse(resp *http.Response, failMsg string) error {
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<14))
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.debugf("response status=%d body=%s", resp.StatusCode, string(data))

	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return errors.New("rate limited by Authelia server, try again later")
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return errors.New(failMsg)
	case resp.StatusCode >= http.StatusBadRequest:
		return fmt.Errorf("%s (status %d)", failMsg, resp.StatusCode)
	}

	var result apiResponse
	if err = json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "OK" {
		return errors.New(failMsg)
	}

	return nil
}
