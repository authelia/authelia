package duo

import (
	"encoding/json"
	"net/url"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

// NewDuoAPI create duo API instance.
func NewDuoAPI(duoAPI *duoapi.DuoApi) *APIImpl {
	return &APIImpl{
		DuoApi: duoAPI,
	}
}

// Call performs a request to the DuoAPI.
func (d *APIImpl) Call(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, values url.Values, method string, path string) (*Response, error) {
	var response Response

	_, responseBytes, err := d.DuoApi.SignedCall(method, path, values)
	if err != nil {
		return nil, err
	}

	ctx.Logger.Tracef("Duo endpoint: %s response raw data for %s from IP %s: %s", path, userSession.Username, ctx.RemoteIP().String(), string(responseBytes))

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}

	if response.Stat == "FAIL" {
		ctx.Logger.Warnf(
			"Duo Push Auth failed to process the auth request for %s from %s: %s (%s), error code %d.",
			userSession.Username, ctx.RemoteIP().String(),
			response.Message, response.MessageDetail, response.Code)
	}

	return &response, nil
}

// PreAuthCall performs a preauth request to the DuoAPI.
func (d *APIImpl) PreAuthCall(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, values url.Values) (*PreAuthResponse, error) {
	var preAuthResponse PreAuthResponse

	response, err := d.Call(ctx, userSession, values, fasthttp.MethodPost, "/auth/v2/preauth")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response.Response, &preAuthResponse)
	if err != nil {
		return nil, err
	}

	return &preAuthResponse, nil
}

// AuthCall performs an auth request to the DuoAPI.
func (d *APIImpl) AuthCall(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, values url.Values) (*AuthResponse, error) {
	var authResponse AuthResponse

	response, err := d.Call(ctx, userSession, values, fasthttp.MethodPost, "/auth/v2/auth")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response.Response, &authResponse)
	if err != nil {
		return nil, err
	}

	return &authResponse, nil
}
