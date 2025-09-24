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
		return nil, err
	}

	ctx.GetLogger().Tracef("Duo endpoint: %s response raw data for %s from IP %s: %s", path, userSession.Username, ctx.RemoteIP().String(), string(body))

	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if response.Stat == "FAIL" {
		ctx.GetLogger().Warnf(
			"Duo Push Auth failed to process the auth request for %s from %s: %s (%s), error code %d.",
			userSession.Username, ctx.RemoteIP().String(),
			response.Message, response.MessageDetail, response.Code)
	}

	return &response, nil
}

func (d *Production) StartupCheck() (err error) {
	var response *http.Response

	if response, _, err = d.Call(fasthttp.MethodGet, "/auth/v2/ping", nil); err != nil {
		return fmt.Errorf("error occurred performing duo ping request: %w", err)
	} else if response.StatusCode != 200 {
		return fmt.Errorf("error occurred performing duo ping request: status code %d", response.StatusCode)
	}

	if response, _, err = d.SignedCall(fasthttp.MethodGet, "/auth/v2/check", nil); err != nil {
		return fmt.Errorf("error occurred performing duo check request: %w", err)
	} else if response.StatusCode != 200 {
		return fmt.Errorf("error occurred performing duo check request: status code %d", response.StatusCode)
	}

	return nil
}

// PreAuthCall performs a preauth request to the DuoAPI.
func (d *Production) PreAuthCall(ctx Context, userSession *session.UserSession, values url.Values) (r *PreAuthResponse, err error) {
	var preAuthResponse PreAuthResponse

	response, err := d.call(ctx, userSession, values, fasthttp.MethodPost, "/auth/v2/preauth")
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(response.Response, &preAuthResponse); err != nil {
		return nil, err
	}

	return &preAuthResponse, nil
}

// AuthCall performs an auth request to the DuoAPI.
func (d *Production) AuthCall(ctx Context, userSession *session.UserSession, values url.Values) (r *AuthResponse, err error) {
	var authResponse AuthResponse

	response, err := d.call(ctx, userSession, values, fasthttp.MethodPost, "/auth/v2/auth")
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(response.Response, &authResponse); err != nil {
		return nil, err
	}

	return &authResponse, nil
}
