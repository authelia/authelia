// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/appauthenticator"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
)

const (
	appErrorInvalidRequest = "invalid_request"
	appErrorBadSignature   = "bad_signature"
	appErrorStaleTimestamp = "stale_timestamp"
	appErrorReplay         = "replay"
	appErrorUnknown        = "unknown"
	appErrorAlreadyUsed    = "already_used"
	appErrorExpired        = "expired"
	appErrorRateLimited    = "rate_limited"
	appErrorInternal       = "internal"

	appResultApproved = "approved"
	appResultDenied   = "denied"
	appResultRejected = "rejected"

	appDecisionApprove = "approve"
	appDecisionDeny    = "deny"

	appCodeChallengeMethodS256 = "S256"

	appFieldDeviceID = "device_id"
	appFieldReason   = "reason"

	appDeviceInfoMaxLength          = 64
	appDevicePlatformMaxLength      = 16
	appAttestationIdentityMaxLength = 255
	appAttestationDetailMaxLength   = 32
	appLastUsedInterval             = time.Minute
	appPullIDAttempts               = 3
)

var appContentTypeJSON = []byte("application/json")

type appErrorResponse struct {
	Error      string `json:"error"`
	ServerTime int64  `json:"server_time"`
}

type appJSONObject map[string]json.RawMessage

// AppAuthenticatorDeviceUnknown answers a request to a device endpoint path which matches no route.
func AppAuthenticatorDeviceUnknown(ctx *middlewares.AutheliaCtx) {
	if appDevicePrecheck(ctx) {
		appWriteError(ctx, fasthttp.StatusNotFound, appErrorUnknown)
	}
}

// AppAuthenticatorDeviceRateLimited answers a rate limited request to a device endpoint.
func AppAuthenticatorDeviceRateLimited(ctx *middlewares.AutheliaCtx, retryAfter time.Duration) {
	ctx.Response.Header.Set(fasthttp.HeaderRetryAfter, strconv.Itoa(int((retryAfter+time.Second-1)/time.Second)))

	appWriteError(ctx, fasthttp.StatusTooManyRequests, appErrorRateLimited)
}

func appServerTime(ctx *middlewares.AutheliaCtx) int64 {
	return ctx.GetClock().Now().Unix()
}

func appWriteJSON(ctx *middlewares.AutheliaCtx, status int, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error occurred encoding the app authenticator response")

		status = fasthttp.StatusInternalServerError
		data, _ = json.Marshal(appErrorResponse{Error: appErrorInternal, ServerTime: appServerTime(ctx)})
	}

	ctx.SetStatusCode(status)
	ctx.SetContentTypeBytes(appContentTypeJSON)
	ctx.SetBody(data)
}

func appWriteError(ctx *middlewares.AutheliaCtx, status int, code string) {
	appWriteJSON(ctx, status, appErrorResponse{Error: code, ServerTime: appServerTime(ctx)})
}

func appWriteInternal(ctx *middlewares.AutheliaCtx, err error, message string) {
	ctx.Logger.WithError(err).Error(message)

	appWriteError(ctx, fasthttp.StatusInternalServerError, appErrorInternal)
}

func appRawTarget(ctx *middlewares.AutheliaCtx) string {
	if uri, ok := ctx.UserValue(middlewares.UserValueKeyRawURI).(string); ok {
		return uri
	}

	return string(ctx.RequestURI())
}

func appRouterPath(ctx *middlewares.AutheliaCtx) string {
	if ctx.Configuration.Server.Address == nil {
		return ""
	}

	if p := ctx.Configuration.Server.Address.RouterPath(); p != "/" {
		return p
	}

	return ""
}

func appDevicePrecheck(ctx *middlewares.AutheliaCtx, params ...string) bool {
	raw := appRawTarget(ctx)

	if (appRouterPath(ctx) != "" && ctx.BasePath() == "") || strings.ContainsAny(raw, "?#") {
		appWriteError(ctx, fasthttp.StatusNotFound, appErrorUnknown)

		return false
	}

	if len(ctx.PostBody()) > appauthenticator.MaxBodySize {
		ctx.SetConnectionClose()

		appWriteError(ctx, fasthttp.StatusBadRequest, appErrorInvalidRequest)

		return false
	}

	if raw != ctx.BasePath()+string(ctx.Path()) {
		appWriteError(ctx, fasthttp.StatusNotFound, appErrorUnknown)

		return false
	}

	for _, param := range params {
		if !appauthenticator.IsValidID(param) {
			appWriteError(ctx, fasthttp.StatusNotFound, appErrorUnknown)

			return false
		}
	}

	return true
}

func appUserValue(ctx *middlewares.AutheliaCtx, key string) string {
	value, _ := ctx.UserValue(key).(string)

	return value
}

func appParseObject(ctx *middlewares.AutheliaCtx) (appJSONObject, bool) {
	var object appJSONObject

	if err := json.Unmarshal(ctx.PostBody(), &object); err != nil || object == nil {
		appWriteError(ctx, fasthttp.StatusBadRequest, appErrorInvalidRequest)

		return nil, false
	}

	return object, true
}

func (o appJSONObject) string(key string) (string, bool) {
	raw, ok := o[key]
	if !ok {
		return "", false
	}

	var value string

	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false
	}

	return value, true
}

func (o appJSONObject) object(key string) (appJSONObject, bool) {
	raw, ok := o[key]
	if !ok {
		return nil, false
	}

	var value appJSONObject

	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		return nil, false
	}

	return value, true
}

func (o appJSONObject) strings(key string) []string {
	raw, ok := o[key]
	if !ok {
		return nil
	}

	var value []string

	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}

	return value
}

func (o appJSONObject) base64(key string) ([]byte, bool) {
	value, ok := o.string(key)
	if !ok || !appauthenticator.IsBase64URL(value) {
		return nil, false
	}

	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, false
	}

	return decoded, true
}

func appCodeChallenge(o appJSONObject) (string, bool) {
	codeChallenge, ok := o.string("code_challenge")
	if !ok || !appauthenticator.IsValidCodeChallenge(codeChallenge) {
		return "", false
	}

	if method, ok := o.string("code_challenge_method"); !ok || method != appCodeChallengeMethodS256 {
		return "", false
	}

	return codeChallenge, true
}

func appTruncate(value string, length int) string {
	if utf8.RuneCountInString(value) <= length {
		return value
	}

	return string([]rune(value)[:length])
}

func appNullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func appSignedURL(ctx *middlewares.AutheliaCtx) (string, error) {
	host := ctx.GetXForwardedHost()
	if len(host) == 0 {
		return "", middlewares.ErrMissingXForwardedHost
	}

	return strings.ToLower(string(ctx.XForwardedProto())) + "://" + strings.ToLower(string(host)) + appRawTarget(ctx), nil
}

func appAuthFailure(ctx *middlewares.AutheliaCtx, code, deviceID, reason string) {
	ctx.Logger.WithFields(map[string]any{appFieldDeviceID: deviceID, appFieldReason: reason}).Infof("App authenticator device request authentication failed with '%s'", code)

	appWriteError(ctx, fasthttp.StatusUnauthorized, code)
}

func appAuthenticate(ctx *middlewares.AutheliaCtx) (device *model.AppAuthenticatorDevice, ok bool) {
	auth, err := appauthenticator.ParseAuthorization(ctx.Request.Header.Peek(fasthttp.HeaderAuthorization))
	if err != nil {
		appAuthFailure(ctx, appErrorBadSignature, "", err.Error())

		return nil, false
	}

	now := ctx.GetClock().Now()

	if skew := now.Unix() - auth.Time.Unix(); skew > int64(appauthenticator.TimestampSkew/time.Second) || -skew > int64(appauthenticator.TimestampSkew/time.Second) {
		appAuthFailure(ctx, appErrorStaleTimestamp, auth.DeviceID, "timestamp is outside the permitted skew")

		return nil, false
	}

	if device, err = ctx.Providers.StorageProvider.LoadAppAuthenticatorDeviceByDeviceID(ctx, auth.DeviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			appAuthFailure(ctx, appErrorBadSignature, auth.DeviceID, "unknown device")
		} else {
			appWriteInternal(ctx, err, "Error occurred loading the app authenticator device")
		}

		return nil, false
	}

	if _, err = base64.RawURLEncoding.DecodeString(auth.Signature); err != nil {
		appAuthFailure(ctx, appErrorBadSignature, auth.DeviceID, "signature is not valid base64url")

		return nil, false
	}

	requestURL, err := appSignedURL(ctx)
	if err != nil {
		appAuthFailure(ctx, appErrorBadSignature, auth.DeviceID, "the request url could not be determined: "+err.Error())

		return nil, false
	}

	transport, err := appauthenticator.ParsePublicKey(device.TransportPublicKey)
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred parsing the stored app authenticator transport key")

		return nil, false
	}

	message := appauthenticator.RequestString(device.BaseURL, auth.DeviceID, string(ctx.Method()), requestURL, auth.Timestamp, auth.Nonce, appauthenticator.BodyHash(ctx.PostBody()))

	if !appauthenticator.VerifySignature(transport, message, auth.Signature) {
		appAuthFailure(ctx, appErrorBadSignature, auth.DeviceID, "bad request signature")

		return nil, false
	}

	inserted, err := ctx.Providers.StorageProvider.SaveAppAuthenticatorRequestNonce(ctx, auth.DeviceID, auth.Nonce, now.Add(appauthenticator.ReplayRetention))
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred saving the app authenticator request nonce")

		return nil, false
	}

	if !inserted {
		appAuthFailure(ctx, appErrorReplay, auth.DeviceID, "request nonce was already used")

		return nil, false
	}

	if !device.LastUsedAt.Valid || now.Sub(device.LastUsedAt.Time) >= appLastUsedInterval {
		if err = ctx.Providers.StorageProvider.UpdateAppAuthenticatorDeviceLastUsed(ctx, device.ID, now); err != nil {
			ctx.Logger.WithError(err).WithField(appFieldDeviceID, device.DeviceID).Error("Error occurred updating the app authenticator device last used time")
		}
	}

	ctx.Logger = ctx.Logger.WithFields(map[string]any{logging.FieldUsername: device.Username, appFieldDeviceID: device.DeviceID})

	return device, true
}

func ctxLogAppAuthenticatorEvent(ctx *middlewares.AutheliaCtx, device *model.AppAuthenticatorDevice, added bool) {
	description, suffix := eventLogAction2FARemoved, eventEmailAction2FARemovedSuffix

	if added {
		description, suffix = eventLogAction2FAAdded, eventEmailAction2FAAddedSuffix
	}

	body := emailEventBody{
		Prefix: eventEmailAction2FAPrefix,
		Body:   eventEmailAction2FABody,
		Suffix: suffix,
	}

	ctxLogEvent(ctx, device.Username, description, body, map[string]any{eventLogKeyAction: description, eventLogKeyCategory: eventLogCategoryAppAuthenticator, eventLogKeyDescription: device.Name})
}
