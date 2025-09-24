// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package duo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// New returns a new Duo [Provider] given the [schema.Configuration], or nil if the Duo API is disabled.
func New(config *schema.Configuration) Provider {
	if config == nil || config.DuoAPI.Disable {
		return nil
	}

	var provider BaseProvider

	if utils.Dev {
		provider = duoapi.NewDuoApi(
			config.DuoAPI.IntegrationKey,
			config.DuoAPI.SecretKey,
			config.DuoAPI.Hostname, "", duoapi.SetInsecure(),
		)
	} else {
		provider = duoapi.NewDuoApi(
			config.DuoAPI.IntegrationKey,
			config.DuoAPI.SecretKey,
			config.DuoAPI.Hostname, "",
		)
	}

	return &Production{BaseProvider: provider}
}

// NewDuoAPI create duo API instance.
func NewDuoAPI(duoAPI BaseProvider) *Production {
	return &Production{
		BaseProvider: duoAPI,
	}
}

func (d *Production) call(ctx Context, userSession *session.UserSession, values url.Values, method string, path string) (r *Response, err error) {
	var (
		response Response
		body     []byte
	)

	if _, body, err = d.SignedCall(method, path, values); err != nil {
		return nil, fmt.Errorf("error occurred making signed call: %w", err)
	}

	ctx.GetLogger().Tracef("Duo endpoint: %s response raw data for %s from IP %s: %s", path, userSession.Username, ctx.RemoteIP().String(), string(body))

	if err = json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error occurred parsing response: %w", err)
	}

	switch response.Stat {
	case "OK":
		ctx.GetLogger().
			WithFields(map[string]any{"status": response.Stat, "message": response.Message, "username": userSession.Username}).
			Trace("Duo Push Auth success response.")

		return &response, nil
	case "FAIL":
		ctx.GetLogger().
			WithFields(map[string]any{"status": response.Stat, "status_code": response.Code, "message": response.Message, "message_detail": response.MessageDetail, "username": userSession.Username}).
			Warn("Duo Push Auth call returned a failure status.")

		return &response, fmt.Errorf("failure status was returned")
	default:
		ctx.GetLogger().
			WithFields(map[string]any{"status": response.Stat, "status_code": response.Code, "message": response.Message, "message_detail": response.MessageDetail, "username": userSession.Username}).
			Warn("Duo Push API call returned an unknown status.")

		return &response, fmt.Errorf("unknown status was returned")
	}
}

// StartupCheck implements the [model.StartupCheck] interface. It performs an unsigned ping request to validate
// connectivity with the Duo API, then a signed check request to validate the configured credentials.
func (d *Production) StartupCheck() (err error) {
	var response *http.Response

	if response, _, err = d.Call(fasthttp.MethodGet, "/auth/v2/ping", nil); err != nil {
		return fmt.Errorf("error occurred performing duo ping request: %w", err)
	} else if response.StatusCode != fasthttp.StatusOK {
		return fmt.Errorf("error occurred performing duo ping request: status code %d", response.StatusCode)
	}

	if response, _, err = d.SignedCall(fasthttp.MethodGet, "/auth/v2/check", nil); err != nil {
		return fmt.Errorf("error occurred performing duo check request: %w", err)
	} else if response.StatusCode != fasthttp.StatusOK {
		return fmt.Errorf("error occurred performing duo check request: status code %d", response.StatusCode)
	}

	return nil
}

// PreAuthCall performs a preauth request to the DuoAPI.
func (d *Production) PreAuthCall(ctx Context, userSession *session.UserSession, values url.Values) (r *PreAuthResponse, err error) {
	var preAuthResponse PreAuthResponse

	response, err := d.call(ctx, userSession, values, fasthttp.MethodPost, "/auth/v2/preauth")
	if err != nil {
		return nil, fmt.Errorf("error occurred making the preauth call to the duo api: %w", err)
	}

	if err = json.Unmarshal(response.Response, &preAuthResponse); err != nil {
		return nil, fmt.Errorf("error occurred parsing the duo api preauth json response: %w", err)
	}

	return &preAuthResponse, nil
}

// AuthCall performs an auth request to the DuoAPI.
func (d *Production) AuthCall(ctx Context, userSession *session.UserSession, values url.Values) (r *AuthResponse, err error) {
	var authResponse AuthResponse

	response, err := d.call(ctx, userSession, values, fasthttp.MethodPost, "/auth/v2/auth")
	if err != nil {
		return nil, fmt.Errorf("error occurred making the auth call to the duo api: %w", err)
	}

	if err = json.Unmarshal(response.Response, &authResponse); err != nil {
		return nil, fmt.Errorf("error occurred parsing the duo api auth json response: %w", err)
	}

	return &authResponse, nil
}
