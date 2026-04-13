package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RFC 8628 token endpoint polling error codes.
const (
	deviceAuthorizationPending = "authorization_pending"
	deviceSlowDown             = "slow_down"
	deviceAccessDenied         = "access_denied"
	deviceExpiredToken         = "expired_token"
)

// DeviceAuthorizationResponse is the RFC 8628 device authorization endpoint response.
type DeviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// DeviceAuthorize initiates an OAuth2 Device Authorization Grant flow.
func (c *AutheliaClient) DeviceAuthorize(clientID, clientSecret, scope string) (*DeviceAuthorizationResponse, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("scope", scope)

	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}

	c.debugf("POST %s/api/oidc/device-authorization", c.baseURL)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/oidc/device-authorization", strings.NewReader(form.Encode())) //nolint:gosec // URL from trusted PAM config.
	if err != nil {
		return nil, fmt.Errorf("failed to create device authorization request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.setHeaders(req)

	resp, err := c.client.Do(req) //nolint:gosec // URL from trusted PAM config.
	if err != nil {
		return nil, fmt.Errorf("device authorization request failed: %w", err)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<14))
	if err != nil {
		return nil, fmt.Errorf("failed to read device authorization response: %w", err)
	}

	c.debugf("device authorization response status=%d body=%s", resp.StatusCode, string(data))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device authorization returned status %d: %s", resp.StatusCode, string(data))
	}

	var result DeviceAuthorizationResponse
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to decode device authorization response: %w", err)
	}

	return &result, nil
}

// DeviceTokenResponse is the token endpoint response for the device flow.
type DeviceTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error"`
}

// devicePollInterval caps the initial token poll interval for UX. RFC 8628 treats the
// server's interval as a minimum; Authelia returns 10 seconds which is unnecessarily
// long. We start shorter and back off to the server's value on slow_down.
const devicePollInterval = 3

// PollDeviceToken polls the token endpoint until the device authorization is approved,
// denied, or expires.
func (c *AutheliaClient) PollDeviceToken(clientID, clientSecret, deviceCode string, expiresIn, interval int) error {
	if interval <= 0 || interval > devicePollInterval {
		interval = devicePollInterval
	}

	deadline := time.Now().Add(time.Duration(expiresIn) * time.Second)

	for first := true; ; first = false {
		if time.Now().After(deadline) {
			return errors.New("device authorization expired before user approval")
		}

		if !first {
			time.Sleep(time.Duration(interval) * time.Second)
		}

		done, slowDown, err := c.pollDeviceTokenOnce(clientID, clientSecret, deviceCode)
		if err != nil {
			return err
		}

		if done {
			return nil
		}

		if slowDown {
			interval += 5
		}
	}
}

// pollDeviceTokenOnce makes a single device token request and classifies the response.
func (c *AutheliaClient) pollDeviceTokenOnce(clientID, clientSecret, deviceCode string) (done, slowDown bool, err error) {
	c.debugf("POST %s/api/oidc/token", c.baseURL)

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	form.Set("client_id", clientID)
	form.Set("device_code", deviceCode)

	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/oidc/token", strings.NewReader(form.Encode())) //nolint:gosec // URL from trusted PAM config.
	if err != nil {
		return false, false, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.setHeaders(req)

	resp, err := c.client.Do(req) //nolint:gosec // URL from trusted PAM config.
	if err != nil {
		return false, false, fmt.Errorf("token request failed: %w", err)
	}

	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<14))
	resp.Body.Close()

	c.debugf("device token response status=%d body=%s", resp.StatusCode, string(data))

	var result DeviceTokenResponse

	_ = json.Unmarshal(data, &result)

	if resp.StatusCode == http.StatusOK && result.AccessToken != "" {
		return true, false, nil
	}

	switch result.Error {
	case deviceAuthorizationPending:
		return false, false, nil
	case deviceSlowDown:
		return false, true, nil
	case deviceAccessDenied:
		return false, false, errors.New("device authorization denied by user")
	case deviceExpiredToken:
		return false, false, errors.New("device authorization token expired")
	default:
		if result.Error != "" {
			return false, false, fmt.Errorf("device token error: %s", result.Error)
		}

		return false, false, fmt.Errorf("device token request returned status %d", resp.StatusCode)
	}
}
