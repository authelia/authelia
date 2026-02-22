package handlers

import (
	"context"
	"errors"
	"net/url"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/authelia/authelia/v4/internal/session"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// Authz is a type which is a effectively is a middlewares.RequestHandler for authorization requests. This should NOT be
// manually used and developers should instead use NewAuthzBuilder.
type Authz struct {
	config AuthzConfig

	strategies []AuthnStrategy

	handleGetObject HandlerAuthzGetObject

	handleGetAutheliaURL HandlerAuthzGetAutheliaURL

	handleAuthorized   HandlerAuthzAuthorized
	handleUnauthorized HandlerAuthzUnauthorized

	implementation AuthzImplementation
}

// HandlerAuthzUnauthorized is a Authz handler func that handles unauthorized responses.
type HandlerAuthzUnauthorized func(ctx AuthzContext, authn *Authn, redirectionURL *url.URL)

// HandlerAuthzAuthorized is a Authz handler func that handles authorized responses.
type HandlerAuthzAuthorized func(ctx AuthzContext, authn *Authn)

// HandlerAuthzGetAutheliaURL is a Authz handler func that handles retrieval of the Portal URL.
type HandlerAuthzGetAutheliaURL func(ctx AuthzContext) (portalURL *url.URL, err error)

// HandlerAuthzGetRedirectionURL is a Authz handler func that handles retrieval of the Redirection URL.
type HandlerAuthzGetRedirectionURL func(ctx AuthzContext, object *authorization.Object) (redirectionURL *url.URL, err error)

// HandlerAuthzGetObject is a Authz handler func that handles retrieval of the authorization.Object to authorize.
type HandlerAuthzGetObject func(ctx AuthzContext) (object authorization.Object, err error)

// HandlerAuthzVerifyObject is a Authz handler func that handles authorization of the authorization.Object.
type HandlerAuthzVerifyObject func(ctx AuthzContext, object authorization.Object) (err error)

// AuthnType is an auth type.
type AuthnType int

const (
	// AuthnTypeNone is a nil Authentication AuthnType.
	AuthnTypeNone AuthnType = iota

	// AuthnTypeCookie is an Authentication AuthnType based on the Cookie header.
	AuthnTypeCookie

	// AuthnTypeProxyAuthorization is an Authentication AuthnType based on the Proxy-Authorization header.
	AuthnTypeProxyAuthorization

	// AuthnTypeAuthorization is an Authentication AuthnType based on the Authorization header.
	AuthnTypeAuthorization
)

// Authn is authentication.
type Authn struct {
	Username string
	Method   string
	ClientID string

	Details authentication.UserDetails
	Level   authentication.Level
	Object  authorization.Object
	Type    AuthnType

	Header HeaderAuthorization
}

type HeaderAuthorization struct {
	Authorization *model.Authorization
	Realm         string
	Scope         string
	Error         *oauthelia2.RFC6749Error
}

// AuthzConfig represents the configuration elements of the Authz type.
type AuthzConfig struct {
	RefreshInterval schema.RefreshIntervalDuration

	// StatusCodeBadRequest is sent for configuration issues prior to performing authorization checks. It's set by the
	// builder.
	StatusCodeBadRequest int
}

// AuthzBuilder is a builder pattern for the Authz type.
type AuthzBuilder struct {
	config         AuthzConfig
	implementation AuthzImplementation
	strategies     []AuthnStrategy
}

// AuthnStrategy is a strategy used for Authz authentication.
type AuthnStrategy interface {
	Get(ctx AuthzContext, manager session.Manager, object *authorization.Object) (authn *Authn, err error)
	CanHandleUnauthorized() (handle bool)
	HeaderStrategy() (is bool)
	HandleUnauthorized(ctx AuthzContext, authn *Authn, redirectionURL *url.URL)
}

// AuthzResult is a result for Authz response handling determination.
type AuthzResult int

const (
	// AuthzResultForbidden means the user is forbidden access to a resource.
	AuthzResultForbidden AuthzResult = iota

	// AuthzResultUnauthorized means the user can access the resource with more permissions.
	AuthzResultUnauthorized

	// AuthzResultAuthorized means the user is authorized given her current permissions.
	AuthzResultAuthorized
)

// AuthzImplementation represents an Authz implementation.
type AuthzImplementation int

// AuthnStrategy names.
const (
	AuthnStrategyCookieSession                       = "CookieSession"
	AuthnStrategyHeaderAuthorization                 = "HeaderAuthorization"
	AuthnStrategyHeaderProxyAuthorization            = "HeaderProxyAuthorization"
	AuthnStrategyHeaderAuthRequestProxyAuthorization = "HeaderAuthRequestProxyAuthorization"
	AuthnStrategyHeaderLegacy                        = "HeaderLegacy"
)

const (
	// AuthzImplLegacy is the legacy Authz implementation (VerifyGET).
	AuthzImplLegacy AuthzImplementation = iota

	// AuthzImplForwardAuth is the modern Forward Auth Authz implementation which is used by Caddy and Traefik.
	AuthzImplForwardAuth

	// AuthzImplAuthRequest is the modern Auth Request Authz implementation which is used by NGINX and modelled after
	// the ingress-nginx k8s ingress.
	AuthzImplAuthRequest

	// AuthzImplExtAuthz is the modern ExtAuthz Authz implementation which is used by Envoy.
	AuthzImplExtAuthz
)

// String returns the text representation of this AuthzImplementation.
func (i AuthzImplementation) String() string {
	switch i {
	case AuthzImplLegacy:
		return "Legacy"
	case AuthzImplForwardAuth:
		return "ForwardAuth"
	case AuthzImplAuthRequest:
		return "AuthRequest"
	case AuthzImplExtAuthz:
		return "ExtAuthz"
	default:
		return ""
	}
}

type AuthzBearerIntrospectionProvider interface {
	GetRegisteredClient(ctx context.Context, id string) (client oidc.Client, err error)
	GetAudienceStrategy(ctx context.Context) (strategy oauthelia2.AudienceMatchingStrategy)
	IntrospectToken(ctx context.Context, token string, tokenUse oauthelia2.TokenUse, session oauthelia2.Session, scope ...string) (oauthelia2.TokenUse, oauthelia2.AccessRequester, error)
}

var (
	errTokenIntent = errors.New("the bearer token doesn't appear to be an authelia bearer token")
)
