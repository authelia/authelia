package middlewares

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/metrics"
	"github.com/authelia/authelia/v4/internal/notification"
	"github.com/authelia/authelia/v4/internal/ntp"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/totp"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

// AutheliaCtx contains all server variables related to Authelia.
type AutheliaCtx struct {
	*fasthttp.RequestCtx

	Logger        *logrus.Entry
	Providers     Providers
	Configuration schema.Configuration

	session *session.Session
}

// Providers contain all provider provided to Authelia.
type Providers struct {
	Authorizer            *authorization.Authorizer
	SessionProvider       *session.Provider
	Regulator             *regulation.Regulator
	OpenIDConnect         *oidc.OpenIDConnectProvider
	Metrics               metrics.Provider
	NTP                   *ntp.Provider
	UserProvider          authentication.UserProvider
	StorageProvider       storage.Provider
	Notifier              notification.Notifier
	Templates             *templates.Provider
	TOTP                  totp.Provider
	PasswordPolicy        PasswordPolicyProvider
	UserAttributeResolver expression.UserAttributeResolver
	MetaDataService       webauthn.MetaDataProvider

	Random random.Provider
	Clock  clock.Provider
}

type Context interface {
	GetClock() (clock clock.Provider)
	GetRandom() (random random.Provider)
	GetLogger() (logger *logrus.Entry)
	GetProviders() (providers Providers)
	GetConfiguration() (config *schema.Configuration)
	RemoteIP() (ip net.IP)

	context.Context
}

type ServiceContext interface {
	GetClock() (clock clock.Provider)
	GetRandom() (random random.Provider)
	GetLogger() (logger *logrus.Entry)
	GetProviders() (providers Providers)
	GetConfiguration() (config *schema.Configuration)

	context.Context
}

// RequestHandler represents an Authelia request handler.
type RequestHandler = func(*AutheliaCtx)

// AutheliaMiddleware represent an Authelia middleware.
type AutheliaMiddleware = func(next RequestHandler) RequestHandler

// Middleware represents a fasthttp middleware.
type Middleware = func(next fasthttp.RequestHandler) (handler fasthttp.RequestHandler)

// Bridge represents the func signature that returns a fasthttp.RequestHandler given a RequestHandler allowing it to
// bridge between the two handlers.
type Bridge = func(RequestHandler) fasthttp.RequestHandler

// BridgeBuilder is used to build a Bridge.
type BridgeBuilder struct {
	config          schema.Configuration
	providers       Providers
	preMiddlewares  []Middleware
	postMiddlewares []AutheliaMiddleware
}

// Basic represents a middleware applied to a fasthttp.RequestHandler.
type Basic func(next fasthttp.RequestHandler) (handler fasthttp.RequestHandler)

// IdentityVerificationStartArgs represent the arguments used to customize the starting phase
// of the identity verification process.
type IdentityVerificationStartArgs struct {
	// Email template needs a subject, a title and the content of the button.
	MailTitle               string
	MailButtonContent       string
	MailButtonRevokeContent string

	// The target endpoint where to redirect the user when verification process
	// is completed successfully.
	TargetEndpoint string

	RevokeEndpoint string

	// The action claim that will be stored in the JWT token.
	ActionClaim string

	// The function retrieving the identity to who the email will be sent.
	IdentityRetrieverFunc func(ctx *AutheliaCtx) (*session.Identity, error)

	// The function for checking the user in the token is valid for the current action.
	IsTokenUserValidFunc func(ctx *AutheliaCtx, username string) bool
}

// IdentityVerificationFinishArgs represent the arguments used to customize the finishing phase
// of the identity verification process.
type IdentityVerificationFinishArgs struct {
	// The action claim that should be in the token to consider the action legitimate.
	ActionClaim string

	// The function for checking the user in the token is valid for the current action.
	IsTokenUserValidFunc func(ctx *AutheliaCtx, username string) bool
}

// IdentityVerificationFinishBody type of the body received by the finish endpoint.
type IdentityVerificationFinishBody struct {
	Token string `json:"token"`
}

// OKResponse model of a status OK response.
type OKResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
}

// ErrorResponse model of an error response.
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// AuthenticationErrorResponse model of an error response.
type AuthenticationErrorResponse struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	Authentication bool   `json:"authentication"`
	Elevation      bool   `json:"elevation"`
}

// ElevatedForbiddenResponse is a response for RequireElevated.
type ElevatedForbiddenResponse struct {
	Elevation    bool `json:"elevation"`
	FirstFactor  bool `json:"first_factor"`
	SecondFactor bool `json:"second_factor"`
}
