package duo

import (
	"encoding/json"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

// NewDuoAPI create duo API instance.
func NewDuoAPI(duoAPI BaseProvider) *Production {
	return &Production{
		BaseProvider: duoAPI,
	}
}

// Call performs a request to the DuoAPI.
func (d *Production) Call(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, values url.Values, method string, path string) (r *Response, err error) {
	var (
		response Response
		body     []byte
	)

	if _, body, err = d.SignedCall(method, path, values); err != nil {
		return nil, err
	}

	ctx.Logger.Tracef("Duo endpoint: %s response raw data for %s from IP %s: %s", path, userSession.Username, ctx.RemoteIP().String(), string(body))

	if err = json.Unmarshal(body, &response); err != nil {
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
func (d *Production) PreAuthCall(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, values url.Values) (r *PreAuthResponse, err error) {
	var preAuthResponse PreAuthResponse

	response, err := d.Call(ctx, userSession, values, fasthttp.MethodPost, "/auth/v2/preauth")
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(response.Response, &preAuthResponse); err != nil {
		return nil, err
	}

	return &preAuthResponse, nil
}

// AuthCall performs an auth request to the DuoAPI.
func (d *Production) AuthCall(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, values url.Values) (r *AuthResponse, err error) {
	var authResponse AuthResponse

	response, err := d.Call(ctx, userSession, values, fasthttp.MethodPost, "/auth/v2/auth")
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(response.Response, &authResponse); err != nil {
		return nil, err
	}

	return &authResponse, nil
}
