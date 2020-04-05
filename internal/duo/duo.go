package duo

import (
	"encoding/json"
	"net/url"

	duoapi "github.com/duosecurity/duo_api_golang"

	"github.com/authelia/authelia/internal/middlewares"
)

// NewDuoAPI create duo API instance
func NewDuoAPI(duoAPI *duoapi.DuoApi) *APIImpl {
	api := new(APIImpl)
	api.DuoApi = duoAPI
	return api
}

// Call call to the DuoAPI
func (d *APIImpl) Call(values url.Values, ctx *middlewares.AutheliaCtx) (*Response, error) {
	_, responseBytes, err := d.DuoApi.SignedCall("POST", "/auth/v2/auth", values)

	if err != nil {
		return nil, err
	}

	ctx.Logger.Tracef("Duo Push Auth Response Raw Data for %s from IP %s: %s", ctx.GetSession().Username, ctx.RemoteIP().String(), string(responseBytes))

	var response Response
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
