package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ResetPasswordDELETE handler for deleting password reset JWT's.
func ResetPasswordDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		token        *jwt.Token
		verification *model.IdentityVerification
		claims       *model.IdentityVerificationClaim
		ok           bool
		err          error
	)

	body := &bodyRequestPasswordResetDELETE{}

	if err = ctx.ParseBody(body); err != nil {
		ctx.Error(fmt.Errorf("error occurred parsing reset password delete body: %w", err), messageOperationFailed)
		return
	}

	token, err = jwt.ParseWithClaims(body.Token, &model.IdentityVerificationClaim{},
		func(token *jwt.Token) (any, error) {
			return []byte(ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret), nil
		},
		jwt.WithIssuedAt(),
		jwt.WithIssuer("Authelia"),
		jwt.WithStrictDecoding(),
		ctx.GetClock().GetJWTWithTimeFuncOption(),
	)

	switch {
	case err == nil:
		break
	case errors.Is(err, jwt.ErrTokenMalformed):
		ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token as it appears to be malformed, this potentially can occur if you've not copied the full link")
		ctx.SetJSONError(messageOperationFailed)

		return
	case errors.Is(err, jwt.ErrTokenExpired):
		ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token validity period as it appears to be expired")
		ctx.SetJSONError(messageOperationFailed)

		return
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token validity period as it appears to only be valid in the future")
		ctx.SetJSONError(messageOperationFailed)

		return
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token signature")
		ctx.SetJSONError(messageOperationFailed)

		return
	default:
		ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if claims, ok = token.Claims.(*model.IdentityVerificationClaim); !ok {
		ctx.Logger.WithError(fmt.Errorf("failed to map the %T claims to a *model.IdentityVerificationClaim", claims)).Error("Error occurred validating the identity verification token claims")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if verification, err = claims.ToIdentityVerification(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token claims as they appear to be malformed")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if verification.Action != ActionResetPassword {
		ctx.Logger.Errorf("Error occurred revoking the identity verification token, the token action '%s' does not match the endpoint action '%s' which is not allowed", claims.Action, ActionResetPassword)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var full *model.IdentityVerification

	if full, err = ctx.Providers.StorageProvider.LoadIdentityVerification(ctx, verification.JTI.String()); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred looking up identity verification during the revocation phase")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if full.RevokedAt.Valid {
		ctx.Logger.Error("Error occurred revoking identity verification token as it's already revoked")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.RevokeIdentityVerification(ctx, verification.JTI.String(), model.NewNullIP(ctx.RemoteIP())); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred revoking identity verification when attempting to save the revocation status to the database")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}

// ResetPasswordPOST handler for resetting passwords.
func ResetPasswordPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Error(fmt.Errorf("error occurred retrieving session for user: %w", err), messageUnableToResetPassword)
		return
	}

	// Those checks unsure that the identity verification process has been initiated and completed successfully
	// otherwise PasswordReset would not be set to true. We can improve the security of this check by making the
	// request expire at some point because here it only expires when the cookie expires.
	if userSession.PasswordResetUsername == nil {
		ctx.Error(fmt.Errorf("no identity verification process has been initiated"), messageUnableToResetPassword)
		return
	}

	username := *userSession.PasswordResetUsername

	var requestBody resetPasswordStep2RequestBody

	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.Error(err, messageUnableToResetPassword)
		return
	}

	if err = ctx.Providers.PasswordPolicy.Check(requestBody.Password); err != nil {
		ctx.Error(err, messagePasswordWeak)
		return
	}

	if err = ctx.Providers.UserProvider.UpdatePassword(username, requestBody.Password); err != nil {
		switch {
		case utils.IsStringInSliceContains(err.Error(), ldapPasswordComplexityCodes),
			utils.IsStringInSliceContains(err.Error(), ldapPasswordComplexityErrors):
			ctx.Error(err, ldapPasswordComplexityCode)
		default:
			ctx.Error(err, messageUnableToResetPassword)
		}

		return
	}

	ctx.Logger.Debugf("Password of user %s has been reset", username)

	// Reset the request.
	userSession.PasswordResetUsername = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Error(fmt.Errorf("unable to update password reset state: %w", err), messageOperationFailed)
		return
	}

	// Send Notification.
	userInfo, err := ctx.Providers.UserProvider.GetDetails(username)
	if err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}

	if len(userInfo.Emails) == 0 {
		ctx.Logger.Error(fmt.Errorf("user %s has no email address configured", username))
		ctx.ReplyOK()

		return
	}

	data := templates.EmailEventValues{
		Title:       "Password changed successfully",
		DisplayName: userInfo.DisplayName,
		RemoteIP:    ctx.RemoteIP().String(),
		Details: map[string]any{
			"Action": "Password Reset",
		},
		BodyPrefix: eventEmailActionPasswordModifyPrefix,
		BodyEvent:  eventEmailActionPasswordReset,
		BodySuffix: eventEmailActionPasswordModifySuffix,
	}

	addresses := userInfo.Addresses()

	ctx.Logger.Debugf("Sending an email to user %s (%s) to inform that the password has changed.",
		username, addresses[0].String())

	if err = ctx.Providers.Notifier.Send(ctx, addresses[0], "Password changed successfully", ctx.Providers.Templates.GetEventEmailTemplate(), data); err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}
}

func identityRetrieverFromStorage(ctx *middlewares.AutheliaCtx) (*session.Identity, error) {
	var requestBody resetPasswordStep1RequestBody

	err := json.Unmarshal(ctx.PostBody(), &requestBody)
	if err != nil {
		return nil, err
	}

	details, err := ctx.Providers.UserProvider.GetDetails(requestBody.Username)
	if err != nil {
		return nil, err
	}

	if len(details.Emails) == 0 {
		return nil, fmt.Errorf("user %s has no email address configured", requestBody.Username)
	}

	return &session.Identity{
		Username:    requestBody.Username,
		DisplayName: details.DisplayName,
		Email:       details.Emails[0],
	}, nil
}

// ResetPasswordIdentityStart is the handler for initiating the identity validation for resetting a password.
// We need to ensure the attacker cannot perform user enumeration by always replying with 200 whatever what happens in backend.
var ResetPasswordIdentityStart = middlewares.IdentityVerificationStart(middlewares.IdentityVerificationStartArgs{
	MailTitle:               "Reset your password",
	MailButtonContent:       "Reset",
	MailButtonRevokeContent: "Revoke",
	TargetEndpoint:          "/reset-password/step2",
	RevokeEndpoint:          "/revoke/reset-password",
	ActionClaim:             ActionResetPassword,
	IdentityRetrieverFunc:   identityRetrieverFromStorage,
}, middlewares.TimingAttackDelay(10, 250, 85, time.Millisecond*500, false))

func resetPasswordIdentityVerificationFinish(ctx *middlewares.AutheliaCtx, username string) {
	var (
		userSession session.UserSession
		err         error
	)

	ctx.ReplyOK()

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to get session to clear password reset flag in session for user '%s'", userSession.Username)

		return
	}

	userSession.PasswordResetUsername = &username

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to clear password reset flag in session for user '%s'", userSession.Username)
	}
}

// ResetPasswordIdentityFinish the handler for finishing the identity validation.
var ResetPasswordIdentityFinish = middlewares.IdentityVerificationFinish(
	middlewares.IdentityVerificationFinishArgs{ActionClaim: ActionResetPassword}, resetPasswordIdentityVerificationFinish)
