package middlewares

import (
	"fmt"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/session"
)

// OTPEscalationProtectedEndpointConfig represents how the Escalation middleware behaves.
type OTPEscalationProtectedEndpointConfig struct {
	Characters                 int
	EmailValidityDuration      time.Duration
	EscalationValidityDuration time.Duration
	Skip2FA                    bool
}

type RequiredLevelProtectedEndpointConfig struct {
	Level authentication.Level
}

type ProtectedEndpointConfig struct {
	OTPEscalation *OTPEscalationProtectedEndpointConfig
	RequiredLevel *RequiredLevelProtectedEndpointConfig
}

func NewProtectedEndpoint(config *ProtectedEndpointConfig) AutheliaMiddleware {
	var handlers []ProtectedEndpointHandler

	if config.RequiredLevel != nil {
		handlers = append(handlers, &RequiredLevelProtectedEndpointHandler{level: config.RequiredLevel.Level})
	}

	if config.OTPEscalation != nil {
		handlers = append(handlers, &OTPEscalationProtectedEndpointHandler{config: config.OTPEscalation})
	}

	return ProtectedEndpoint(handlers...)
}

func ProtectedEndpoint(handlers ...ProtectedEndpointHandler) AutheliaMiddleware {
	n := len(handlers)

	return func(next RequestHandler) RequestHandler {
		return func(ctx *AutheliaCtx) {
			session, err := ctx.GetSession()

			if err != nil || session.IsAnonymous() {
				ctx.SetAuthenticationErrorJSON(fasthttp.StatusUnauthorized, fasthttp.StatusMessage(fasthttp.StatusUnauthorized), false, false)

				return
			}

			var failed, failedAuthentication, failedElevation bool

			for i := 0; i < n; i++ {
				if handlers[i].Check(ctx, &session) {
					continue
				}

				failed = true

				if handlers[i].IsAuthentication() {
					failedAuthentication = true
				}

				if handlers[i].IsElevation() {
					failedElevation = true
				}

				handlers[i].Failure(ctx, &session)
			}

			if failed {
				ctx.SetAuthenticationErrorJSON(fasthttp.StatusForbidden, fasthttp.StatusMessage(fasthttp.StatusForbidden), failedAuthentication, failedElevation)

				return
			}

			next(ctx)
		}
	}
}

type ProtectedEndpointHandler interface {
	Name() string
	Check(ctx *AutheliaCtx, s *session.UserSession) (success bool)
	Failure(ctx *AutheliaCtx, s *session.UserSession)

	IsAuthentication() bool
	IsElevation() bool
}

func NewRequiredLevelProtectedEndpointHandler(level authentication.Level, statusCode int) *RequiredLevelProtectedEndpointHandler {
	handler := &RequiredLevelProtectedEndpointHandler{
		level:      level,
		statusCode: statusCode,
	}

	if handler.statusCode == 0 {
		handler.statusCode = fasthttp.StatusForbidden
	}

	if handler.level == 0 {
		handler.level = authentication.OneFactor
	}

	return handler
}

type RequiredLevelProtectedEndpointHandler struct {
	level      authentication.Level
	statusCode int
}

func (h *RequiredLevelProtectedEndpointHandler) Name() string {
	return fmt.Sprintf("required_level(%s)", h.level)
}

func (h *RequiredLevelProtectedEndpointHandler) IsAuthentication() bool {
	return true
}

func (h *RequiredLevelProtectedEndpointHandler) IsElevation() bool {
	return false
}

func (h *RequiredLevelProtectedEndpointHandler) Check(ctx *AutheliaCtx, s *session.UserSession) (success bool) {
	return s.AuthenticationLevel >= h.level
}

func (h *RequiredLevelProtectedEndpointHandler) Failure(_ *AutheliaCtx, _ *session.UserSession) {
}

func NewOTPEscalationProtectedEndpointHandler(config OTPEscalationProtectedEndpointConfig) *OTPEscalationProtectedEndpointHandler {
	return &OTPEscalationProtectedEndpointHandler{
		config: &config,
	}
}

type OTPEscalationProtectedEndpointHandler struct {
	config *OTPEscalationProtectedEndpointConfig
}

func (h *OTPEscalationProtectedEndpointHandler) Name() string {
	return "one_time_password"
}

func (h *OTPEscalationProtectedEndpointHandler) IsAuthentication() bool {
	return false
}

func (h *OTPEscalationProtectedEndpointHandler) IsElevation() bool {
	return true
}

func (h *OTPEscalationProtectedEndpointHandler) Check(ctx *AutheliaCtx, s *session.UserSession) (success bool) {
	if h.config.Skip2FA && s.AuthenticationLevel >= authentication.TwoFactor {
		ctx.Logger.
			WithField("username", s.Username).
			Warning("User elevated session check has skipped due to 2FA")

		return true
	}

	if s.Elevations.User == nil {
		ctx.Logger.
			WithField("username", s.Username).
			Warning("User elevated session has not been created")

		return false
	}

	if s.Elevations.User.Expires.Before(ctx.Clock.Now()) {
		ctx.Logger.
			WithField("username", s.Username).
			WithField("expires", s.Elevations.User.Expires).
			Debug("User elevated session IP did not match the request")

		return false
	}

	if !ctx.RemoteIP().Equal(s.Elevations.User.RemoteIP) {
		ctx.Logger.
			WithField("username", s.Username).
			WithField("elevation_ip", s.Elevations.User.RemoteIP).
			Warning("User elevated session IP did not match the request")

		return false
	}

	return true
}

func (h *OTPEscalationProtectedEndpointHandler) Failure(ctx *AutheliaCtx, s *session.UserSession) {
	if s.Elevations.User != nil {
		// If we make it here we should destroy the elevation data.
		s.Elevations.User = nil

		if err := ctx.SaveSession(*s); err != nil {
			ctx.Logger.WithError(err).Error("Error session after user elevated session failure")
		}
	}
}

// Require1FA requires the user to have authenticated with at least one-factor authentication (i.e. password).
func Require1FA(next RequestHandler) RequestHandler {
	handler := ProtectedEndpoint(NewRequiredLevelProtectedEndpointHandler(authentication.OneFactor, fasthttp.StatusForbidden))

	return handler(next)
}

// Require2FA requires the user to have authenticated with two-factor authentication.
func Require2FA(next RequestHandler) RequestHandler {
	handler := ProtectedEndpoint(NewRequiredLevelProtectedEndpointHandler(authentication.TwoFactor, fasthttp.StatusForbidden))

	return handler(next)
}

// Require2FAWithAPIResponse requires the user to have authenticated with two-factor authentication.
func Require2FAWithAPIResponse(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		session, err := ctx.GetSession()

		if err != nil || session.AuthenticationLevel < authentication.TwoFactor {
			ctx.SetAuthenticationErrorJSON(fasthttp.StatusForbidden, "Authentication Required.", true, false)
			return
		}

		next(ctx)
	}
}
