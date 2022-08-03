package handlers

import (
	"net/url"
	"time"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// Authz is a type which is a effectively is a middlewares.RequestHandler for authorization requests.
type Authz struct {
	config AuthzConfig

	fObjectGet    func(ctx *middlewares.AutheliaCtx) (object authorization.Object, err error)
	fObjectVerify func(ctx *middlewares.AutheliaCtx, object authorization.Object) (err error)

	strategies []AuthnStrategy

	fHandleAuthorized   func(ctx *middlewares.AutheliaCtx, authn *Authn)
	fHandleUnauthorized AuthzUnauthorizedHandler
}

// AuthzConfig represents the configuration elements of the Authz type.
type AuthzConfig struct {
	RefreshInterval time.Duration
	Domains         []AuthzDomain
}

// AuthzDomain represents a domain for the AuthzConfig.
type AuthzDomain struct {
	Name      string
	PortalURL *url.URL
}

// AuthzBuilder is a builder pattern for the Authz type.
type AuthzBuilder struct {
	config     AuthzConfig
	impl       AuthzImplementation
	strategies []AuthnStrategy
}

// AuthnStrategy is a strategy used for Authz authentication.
type AuthnStrategy interface {
	Get(ctx *middlewares.AutheliaCtx) (authn Authn, err error)
	CanHandleUnauthorized() (handle bool)
	HandleUnauthorized(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL)
}

// AuthzUnauthorizedHandler is a Authz handler func that handles unauthorized responses.
type AuthzUnauthorizedHandler func(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL)

// AuthzResult is a result for Authz response handling determination.
type AuthzResult int

const (
	// AuthzResultForbidden means the user is forbidden the access to a resource.
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
	default:
		return ""
	}
}
