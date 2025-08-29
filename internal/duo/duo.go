package duo

import (
	"encoding/json"
	"fmt"
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
func (d *Production) Call(ctx middlewares.Context, userSession *session.UserSession, values url.Values, method string, path string) (r *Response, err error) {
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
	// The status string OK is the only one that are expected to be returned with a response.
	case "OK":
		switch response.Code {
		// The status codes 200 and 404 are the only status codes that are expected to be returned with a response.
		case fasthttp.StatusOK, fasthttp.StatusNotFound:
			ctx.GetLogger().
				WithFields(map[string]any{"status": response.Stat, "status_code": response.Code, "message": response.Message, "message_detail": response.MessageDetail, "username": userSession.Username}).
				Trace("Duo Push Auth success response.")

			return &response, nil
		default:
			ctx.GetLogger().
				WithFields(map[string]any{"status": response.Stat, "status_code": response.Code, "message": response.Message, "message_detail": response.MessageDetail, "username": userSession.Username}).
				Warn("Duo Push Auth call returned a failure status code.")

			return &response, fmt.Errorf("failure status code was returned")
		}
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

// PreAuthCall performs a preauth request to the DuoAPI.
func (d *Production) PreAuthCall(ctx middlewares.Context, userSession *session.UserSession, values url.Values) (r *PreAuthResponse, err error) {
	var preAuthResponse PreAuthResponse

	response, err := d.Call(ctx, userSession, values, fasthttp.MethodPost, "/auth/v2/preauth")
	if err != nil {
		return nil, fmt.Errorf("error occurred making the preauth call to the duo api: %w", err)
	}

	if err = json.Unmarshal(response.Response, &preAuthResponse); err != nil {
		return nil, fmt.Errorf("error occurred parsing the duo api preauth json response: %w", err)
	}

	return &preAuthResponse, nil
}

// AuthCall performs an auth request to the DuoAPI.
func (d *Production) AuthCall(ctx middlewares.Context, userSession *session.UserSession, values url.Values) (r *AuthResponse, err error) {
	var authResponse AuthResponse

	response, err := d.Call(ctx, userSession, values, fasthttp.MethodPost, "/auth/v2/auth")
	if err != nil {
		return nil, fmt.Errorf("error occurred making the auth call to the duo api: %w", err)
	}

	if err = json.Unmarshal(response.Response, &authResponse); err != nil {
		return nil, fmt.Errorf("error occurred parsing the duo api auth json response: %w", err)
	}

	return &authResponse, nil
}
