package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/duo-labs/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/models"
	"github.com/authelia/authelia/v4/internal/session"
)

func getWebAuthnUser(ctx *middlewares.AutheliaCtx, userSession session.UserSession) (user *models.WebauthnUser, err error) {
	user = &models.WebauthnUser{
		Username:    userSession.Username,
		DisplayName: userSession.DisplayName,
	}

	if user.Devices, err = ctx.Providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, userSession.Username); err != nil {
		return nil, err
	}

	return user, nil
}

func getWebauthn(ctx *middlewares.AutheliaCtx) (w *webauthn.WebAuthn, appid string, err error) {
	var (
		headerProtoV, headerXForwardedHostV, headerOriginV []byte
		originURL                                          *url.URL
	)

	if headerProtoV = ctx.XForwardedProto(); headerProtoV == nil {
		return nil, "", errMissingXForwardedProto
	}

	if headerXForwardedHostV = ctx.XForwardedHost(); headerXForwardedHostV == nil {
		return nil, "", errMissingXForwardedHost
	}

	appid = fmt.Sprintf("%s://%s", headerProtoV, headerXForwardedHostV)

	if headerOriginV = ctx.Request.Header.PeekBytes(headerOrigin); headerOriginV == nil {
		return nil, "", errMissingOrigin
	}

	if originURL, err = url.Parse(string(headerOriginV)); err != nil {
		return nil, "", fmt.Errorf(errFmtInvalidOrigin, headerOriginV, err)
	}

	if strings.EqualFold(originURL.Scheme, string(headerProtoV)) {
		return nil, "", fmt.Errorf(errFmtInvalidOrigin, headerOriginV, fmt.Errorf(errFmtOriginProtoMismatch, headerProtoV))
	}

	if strings.EqualFold(originURL.Host, string(headerXForwardedHostV)) {
		return nil, "", fmt.Errorf(errFmtInvalidOrigin, headerOriginV, fmt.Errorf(errFmtOriginHostMismatch, headerXForwardedHostV))
	}

	config := &webauthn.Config{
		RPDisplayName: "Authelia",
		RPID:          appid,
		RPOrigin:      string(headerOriginV),

		AttestationPreference: ctx.Configuration.Webauthn.AttestationPreference,
		Timeout:               ctx.Configuration.Webauthn.Timeout,
		Debug:                 ctx.Configuration.Webauthn.Debug,
	}

	ctx.Logger.Tracef("Creating new Webauthn RP instance with ID %s and Origin %s", appid, string(headerOriginV))

	w, err = webauthn.New(config)

	return w, appid, err
}
