package handlers

import (
	"fmt"
	"regexp"
	"time"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
)

var reTokenFormat = regexp.MustCompile(`^[a-zA-Z0-9]{32}$`)

// TelegramAuthRequestGET initiates a Telegram auth request.
func TelegramAuthRequestGET(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session for Telegram auth")
		ctx.ReplyForbidden()

		return
	}

	if userSession.Username == "" {
		ctx.ReplyForbidden()

		return
	}

	// Clean up: delete any pending (unverified) tokens for this user to prevent DoS
	_ = ctx.Providers.StorageProvider.DeleteTelegramVerificationsPending(ctx, userSession.Username)

	// Also purge globally expired tokens
	ttl := 300
	if ctx.Configuration.Telegram != nil && ctx.Configuration.Telegram.TokenTTL > 0 {
		ttl = ctx.Configuration.Telegram.TokenTTL
	}

	_ = ctx.Providers.StorageProvider.DeleteTelegramVerificationsExpired(ctx, time.Now().Add(-time.Duration(ttl)*time.Second))

	// Generate a verification token and store it
	token := ctx.Providers.Random.StringCustom(32, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	err = ctx.Providers.StorageProvider.SaveTelegramVerification(ctx, model.TelegramVerification{
		Username: userSession.Username,
		Token:    token,
	})
	if err != nil {
		ctx.Logger.WithError(err).Error("Error saving Telegram verification token")
		ctx.ReplyStatusCode(500)

		return
	}

	botUsername := ""
	if ctx.Configuration.Telegram != nil {
		botUsername = ctx.Configuration.Telegram.BotUsername
	}

	if err = ctx.SetJSONBody(bodyTelegramAuthRequestResponse{
		Token:       token,
		BotUsername:  botUsername,
		BotDeepLink: fmt.Sprintf("https://t.me/%s?start=%s", botUsername, token),
	}); err != nil {
		ctx.Logger.WithError(err).Error("Error setting JSON body for Telegram auth request")
		ctx.ReplyStatusCode(500)
	}
}

// TelegramAuthStatusGET checks the verification status of a Telegram auth token.
func TelegramAuthStatusGET(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.ReplyForbidden()

		return
	}

	if userSession.Username == "" {
		ctx.ReplyForbidden()

		return
	}

	token := ctx.UserValue("token")
	if token == nil {
		ctx.ReplyBadRequest()

		return
	}

	tokenStr := fmt.Sprintf("%s", token)

	// Validate token format
	if !reTokenFormat.MatchString(tokenStr) {
		ctx.ReplyBadRequest()

		return
	}

	// Enforce TTL
	ttl := 300
	if ctx.Configuration.Telegram != nil && ctx.Configuration.Telegram.TokenTTL > 0 {
		ttl = ctx.Configuration.Telegram.TokenTTL
	}

	verification, err := ctx.Providers.StorageProvider.LoadTelegramVerification(ctx, userSession.Username, tokenStr, time.Now().Add(-time.Duration(ttl)*time.Second))
	if err != nil || verification == nil {
		if err = ctx.SetJSONBody(bodyTelegramAuthStatusResponse{
			Verified: false,
			Expired:  true,
		}); err != nil {
			ctx.Logger.WithError(err).Error("Error setting JSON body")
		}

		return
	}

	// Return only verified boolean — no PII (phone/telegram_id) exposed
	if err = ctx.SetJSONBody(bodyTelegramAuthStatusResponse{
		Verified: verification.Verified,
		Expired:  false,
	}); err != nil {
		ctx.Logger.WithError(err).Error("Error setting JSON body")
	}
}

// TelegramAuthCompletePOST completes the Telegram 2FA after phone verification.
func TelegramAuthCompletePOST(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")
		ctx.ReplyForbidden()

		return
	}

	if userSession.Username == "" {
		ctx.ReplyForbidden()

		return
	}

	bodyJSON := bodySignTelegramRequest{}
	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred parsing body")
		ctx.ReplyBadRequest()

		return
	}

	// Validate token format
	if !reTokenFormat.MatchString(bodyJSON.Token) {
		ctx.ReplyBadRequest()

		return
	}

	// Enforce TTL
	ttl := 300
	if ctx.Configuration.Telegram != nil && ctx.Configuration.Telegram.TokenTTL > 0 {
		ttl = ctx.Configuration.Telegram.TokenTTL
	}

	// Load and verify the token (with expiry)
	verification, err := ctx.Providers.StorageProvider.LoadTelegramVerification(ctx, userSession.Username, bodyJSON.Token, time.Now().Add(-time.Duration(ttl)*time.Second))
	if err != nil || verification == nil || !verification.Verified {
		ctx.Logger.Errorf("Telegram verification not found or not verified for user %s", userSession.Username)

		ctx.Providers.Regulator.HandleAttempt(ctx, false, false, userSession.Username, "", "", regulation.AuthTypeTelegram)

		ctx.ReplyUnauthorized()

		return
	}

	ctx.Providers.Regulator.HandleAttempt(ctx, true, false, userSession.Username, "", "", regulation.AuthTypeTelegram)

	userSession.SetTwoFactorTelegram(ctx.GetClock().Now())

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred saving session")
		ctx.ReplyStatusCode(500)

		return
	}

	// Clean up the verification token
	if err = ctx.Providers.StorageProvider.DeleteTelegramVerification(ctx, userSession.Username, bodyJSON.Token); err != nil {
		ctx.Logger.WithError(err).Warn("Failed to delete telegram verification token")
	}

	Handle2FAResponse(ctx, bodyJSON.TargetURL)
}
