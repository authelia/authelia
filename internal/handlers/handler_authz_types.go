package handlers

import (
	"net/url"
	"time"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

type AuthzBuilder struct {
	config     AuthzConfig
	impl       AuthzImplementation
	strategies []AuthnStrategy
}

type AuthzConfig struct {
	RefreshInterval time.Duration
	Domains         []AuthzDomain
}

type AuthzDomain struct {
	Name      string
	PortalURL *url.URL
}

type AuthzUnauthorizedHandler func(ctx *middlewares.AutheliaCtx, authn *Authn, redirectionURL *url.URL)

type Authz struct {
	config AuthzConfig

	fObjectGet    func(ctx *middlewares.AutheliaCtx) (object authorization.Object, err error)
	fObjectVerify func(ctx *middlewares.AutheliaCtx, object authorization.Object) (err error)

	strategies []AuthnStrategy

	fHandleAuthorized   func(ctx *middlewares.AutheliaCtx, authn *Authn)
	fHandleUnauthorized AuthzUnauthorizedHandler
}

// AuthzImplementation represents an Authz implementation.
type AuthzImplementation int

const (
	// AuthzImplLegacy is the legacy Authz implementation (VerifyGET).
	AuthzImplLegacy AuthzImplementation = iota

	// AuthzImplForwardAuth is the modern Forward Auth Authz implementation which is used by Caddy and Traefik.
	AuthzImplForwardAuth

	// AuthzImplAuthRequest is the modern Auth Request Authz implementation which is used by NGINX and modelled after
	// the ingress-nginx k8s ingress.
	AuthzImplAuthRequest
)
