package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TOTPRegisterGET(ctx *middlewares.AutheliaCtx) {
	if err := ctx.SetJSONBody(ctx.Providers.TOTP.Options()); err != nil {
		ctx.Logger.Errorf("Unable to set TOTP options response in body: %s", err)
	}
}

func TOTPRegisterPUT(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		bodyJSON    bodyRegisterTOTP
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred unmarshaling body %s registration", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	opts := ctx.Providers.TOTP.Options()

	var hasAlgorithm, hasLength, hasPeriod bool

	hasAlgorithm = utils.IsStringInSlice(bodyJSON.Algorithm, opts.Algorithms)

	for _, period := range opts.Periods {
		if period == bodyJSON.Period {
			hasPeriod = true
			break
		}
	}

	for _, length := range opts.Lengths {
		if length == bodyJSON.Length {
			hasLength = true
			break
		}
	}

	if !hasAlgorithm || !hasPeriod || !hasLength {
		ctx.Logger.Errorf("Validation failed for %s registration because the input options were not permitted by the configuration", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	var config *model.TOTPConfiguration

	if config, err = ctx.Providers.TOTP.GenerateCustom(userSession.Username, bodyJSON.Algorithm, "", uint(bodyJSON.Length), uint(bodyJSON.Period), 0); err != nil {
		ctx.Error(fmt.Errorf("unable to generate TOTP key: %w", err), messageUnableToRegisterOneTimePassword)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	userSession.TOTP = &session.TOTP{
		Issuer:    config.Issuer,
		Algorithm: config.Algorithm,
		Digits:    config.Digits,
		Period:    config.Period,
		Secret:    string(config.Secret),
		Expires:   ctx.Clock.Now().Add(time.Minute * 10),
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Error(err, messageUnableToRegisterOneTimePassword)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	response := TOTPKeyResponse{
		OTPAuthURL:   config.URI(),
		Base32Secret: userSession.TOTP.Secret,
	}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Logger.Errorf("Unable to set TOTP key response in body: %s", err)
	}
}

func TOTPRegisterPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		bodyJSON    bodyRegisterFinishTOTP
		valid       bool
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	if userSession.TOTP == nil {
		ctx.Logger.Errorf("Error occurred during %s registration: the user did not initiate a registration on their current session", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	if ctx.Clock.Now().After(userSession.TOTP.Expires) {
		ctx.Logger.Errorf("Error occurred during %s registration: the registration is expired", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	if err = json.Unmarshal(ctx.PostBody(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred unmarshaling body %s registration", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	config := model.TOTPConfiguration{
		CreatedAt: ctx.Clock.Now(),
		Username:  userSession.Username,
		Issuer:    userSession.TOTP.Issuer,
		Algorithm: userSession.TOTP.Algorithm,
		Period:    userSession.TOTP.Period,
		Digits:    userSession.TOTP.Digits,
		Secret:    []byte(userSession.TOTP.Secret),
	}

	if valid, err = ctx.Providers.TOTP.Validate(bodyJSON.Token, &config); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating %s registration", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	} else if !valid {
		ctx.Logger.Errorf("Error occurred validating %s registration", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveTOTPConfiguration(ctx, config); err != nil {
		ctx.Logger.Errorf("Error occurred saving %s registration", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	userSession.TOTP = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf("Error occurred saving session during %s registration", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}
}

func TOTPRegisterDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration cancel", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}

	if userSession.TOTP == nil {
		return
	}

	userSession.TOTP = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf("Error occurred saving session during %s registration cancel", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToRegisterOneTimePassword)

		return
	}
}

func TOTPConfigurationDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration cancel", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToDeleteOneTimePassword)

		return
	}

	if _, err = ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration cancel", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToDeleteOneTimePassword)

		return
	}

	if err = ctx.Providers.StorageProvider.DeleteTOTPConfiguration(ctx, userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving session for %s registration cancel", regulation.AuthTypeTOTP)

		respondUnauthorized(ctx, messageUnableToDeleteOneTimePassword)

		return
	}
}
