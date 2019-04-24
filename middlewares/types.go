package middlewares

import (
	"github.com/clems4ever/authelia/authentication"
	"github.com/clems4ever/authelia/authorization"
	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/clems4ever/authelia/notification"
	"github.com/clems4ever/authelia/regulation"
	"github.com/clems4ever/authelia/session"
	"github.com/clems4ever/authelia/storage"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

// AutheliaCtx contains all server variables related to Authelia.
type AutheliaCtx struct {
	*fasthttp.RequestCtx

	Logger        *logrus.Entry
	Providers     Providers
	Configuration schema.Configuration
	userSession   session.UserSession
}

// Providers contain all provider provided to Authelia.
type Providers struct {
	Authorizer      *authorization.Authorizer
	SessionProvider *session.Provider
	Regulator       *regulation.Regulator

	UserProvider    authentication.UserProvider
	StorageProvider storage.Provider
	Notifier        notification.Notifier
}

// RequestHandler represents an Authelia request handler.
type RequestHandler = func(*AutheliaCtx)

// Middleware represent an Authelia middleware.
type Middleware = func(RequestHandler) RequestHandler

// IdentityVerificationStartArgs represent the arguments used to customize the starting phase
// of the identity verification process.
type IdentityVerificationStartArgs struct {
	// Email template needs a subject, a title and the content of the button.
	MailSubject       string
	MailTitle         string
	MailButtonContent string

	// The target endpoint where to redirect the user when verification process
	// is completed successfully.
	TargetEndpoint string

	// The action claim that will be stored in the JWT token
	ActionClaim string

	// The function retrieving the identity to who the email will be sent.
	IdentityRetrieverFunc func(ctx *AutheliaCtx) (*session.Identity, error)

	// The function for checking the user in the token is valid for the current action
	IsTokenUserValidFunc func(ctx *AutheliaCtx, username string) bool
}

// IdentityVerificationFinishArgs represent the arguments used to customize the finishing phase
// of the identity verification process.
type IdentityVerificationFinishArgs struct {
	// The action claim that should be in the token to consider the action legitimate
	ActionClaim string

	// The function for checking the user in the token is valid for the current action
	IsTokenUserValidFunc func(ctx *AutheliaCtx, username string) bool
}

// IdentityVerificationClaim custom claim for specifying the action claim.
// The action can be to register a TOTP device, a U2F device or reset one's password.
type IdentityVerificationClaim struct {
	jwt.StandardClaims

	// The action this token has been crafted for.
	Action string `json:"action"`
	// The user this token has been crafted for.
	Username string `json:"username"`
}

// IdentityVerificationFinishBody type of the body received by the finish endpoint.
type IdentityVerificationFinishBody struct {
	Token string `json:"token"`
}

// OKResponse model  of a status OK response
type OKResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

// ErrorResponse model of an error response
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
