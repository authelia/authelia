// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/idjag"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/model"
)

// NewIDJAGConfig returns the runtime Identity Assertion JWT Authorization Grant configuration given the provider
// configuration. The configuration is expected to have been validated.
func NewIDJAGConfig(config *schema.IdentityProvidersOpenIDConnect) (c *IDJAGConfig, err error) {
	grant := config.IdentityAssertionAuthorizationGrant

	c = &IDJAGConfig{
		Lifespan:                     grant.Lifespan,
		SingleUse:                    grant.SingleUse == nil || *grant.SingleUse,
		AccessTokenAllowance:         grant.AccessTokenAllowance,
		ResourceOptional:             grant.ResourceOptional,
		ResourceAuthorizationServers: make(map[string]*IDJAGResourceAuthorizationServer, len(grant.ResourceAuthorizationServers)),
		Aliases:                      map[string]string{},
		TrustedIssuers:               make(map[string]*IDJAGTrustedIssuer, len(grant.TrustedIssuers)),
	}

	if c.Lifespan <= 0 {
		c.Lifespan = idjagLifespanDefault
	}

	var scopes []string

	if len(grant.ResourceAuthorizationServers) != 0 {
		scopes = make([]string, 0, len(config.Scopes))

		for scope := range config.Scopes {
			scopes = append(scopes, scope)
		}
	}

	for _, server := range grant.ResourceAuthorizationServers {
		if server.Issuer == nil {
			continue
		}

		ras := &IDJAGResourceAuthorizationServer{
			Issuer:           server.Issuer.String(),
			SigningAlg:       server.SigningAlgorithm,
			MaxAge:           server.MaxAge,
			SubjectClientID:  server.SubjectClientID,
			ClaimsPolicy:     server.ClaimsPolicy,
			JWTDPoPSupported: server.JWTDPoPSupported,
		}

		if ras.SigningAlg == "" {
			ras.SigningAlg = SigningAlgRSAUsingSHA256
		}

		if ras.ClaimsPolicy != "" {
			ras.claims = NewCustomClaimsStrategy(ras.ClaimsPolicy, scopes, config.Scopes, config.ClaimsPolicies)
		}

		c.ResourceAuthorizationServers[ras.Issuer] = ras

		for _, alias := range server.Aliases {
			c.Aliases[alias] = ras.Issuer
		}
	}

	for _, issuer := range grant.TrustedIssuers {
		if issuer.Issuer == nil {
			continue
		}

		trusted := &IDJAGTrustedIssuer{
			IDJAGTrustedIssuer: oauthelia2.IDJAGTrustedIssuer{
				Issuer:      issuer.Issuer.String(),
				JSONWebKeys: NewJSONWebKeySet(issuer.JSONWebKeys),
				SigningAlgs: issuer.SigningAlgorithms,
				Clients:     issuer.Clients,
			},
			Resources:     issuer.Resources,
			UsernameClaim: issuer.UsernameClaim,
		}

		if issuer.JSONWebKeysURI != nil {
			trusted.JSONWebKeysURI = issuer.JSONWebKeysURI.String()
		}

		if len(trusted.SigningAlgs) == 0 {
			trusted.SigningAlgs = []string{SigningAlgRSAUsingSHA256}
		}

		if trusted.AMR, err = expression.NewAuthenticationMethodsReferencesMapping(issuer.AuthenticationMethodsReferences); err != nil {
			return nil, fmt.Errorf("error occurred compiling the authentication methods references expression for the trusted issuer '%s': %w", trusted.Issuer, err)
		}

		c.TrustedIssuers[trusted.Issuer] = trusted
	}

	c.issueSigningAlgs, c.redeemSigningAlgs = c.signingAlgs()

	return c, nil
}

func (c *IDJAGConfig) signingAlgs() (issue, redeem []string) {
	if len(c.ResourceAuthorizationServers) != 0 {
		issue = []string{SigningAlgRSAUsingSHA256}

		for _, ras := range c.ResourceAuthorizationServers {
			issue = append(issue, ras.SigningAlg)
		}

		issue = sortedUnique(issue)
	}

	if len(c.TrustedIssuers) != 0 {
		for _, trusted := range c.TrustedIssuers {
			redeem = append(redeem, trusted.SigningAlgs...)
		}

		redeem = sortedUnique(redeem)
	}

	return issue, redeem
}

// IDJAGConfig is the runtime Identity Assertion JWT Authorization Grant configuration.
type IDJAGConfig struct {
	Lifespan             time.Duration
	AccessTokenAllowance time.Duration
	SingleUse            bool
	ResourceOptional     bool

	// ResourceAuthorizationServers is keyed by the issuer of the Resource Authorization Server.
	ResourceAuthorizationServers map[string]*IDJAGResourceAuthorizationServer

	// Aliases maps an alias to the issuer of the Resource Authorization Server it identifies.
	Aliases map[string]string

	// TrustedIssuers is keyed by the issuer.
	TrustedIssuers map[string]*IDJAGTrustedIssuer

	issueSigningAlgs  []string
	redeemSigningAlgs []string
}

// LookupResourceAuthorizationServer returns the Resource Authorization Server identified by the audience, matching the
// issuer before the aliases.
func (c *IDJAGConfig) LookupResourceAuthorizationServer(audience string) (ras *IDJAGResourceAuthorizationServer, ok bool) {
	if c == nil {
		return nil, false
	}

	if ras, ok = c.ResourceAuthorizationServers[audience]; ok {
		return ras, true
	}

	var issuer string

	if issuer, ok = c.Aliases[audience]; !ok {
		return nil, false
	}

	if ras, ok = c.ResourceAuthorizationServers[issuer]; !ok {
		return nil, false
	}

	return ras, true
}

// LookupTrustedIssuer returns the trusted issuer with the given issuer identifier.
func (c *IDJAGConfig) LookupTrustedIssuer(issuer string) (trusted *IDJAGTrustedIssuer, ok bool) {
	if c == nil {
		return nil, false
	}

	trusted, ok = c.TrustedIssuers[issuer]

	return trusted, ok
}

// IDJAGResourceAuthorizationServer is a Resource Authorization Server Identity Assertion JWT Authorization Grants may
// be issued for.
type IDJAGResourceAuthorizationServer struct {
	Issuer           string
	SigningAlg       string
	MaxAge           time.Duration
	SubjectClientID  string
	ClaimsPolicy     string
	JWTDPoPSupported bool

	claims *CustomClaimsStrategy
}

// IDJAGTrustedIssuer is an issuer whose Identity Assertion JWT Authorization Grants are accepted.
type IDJAGTrustedIssuer struct {
	oauthelia2.IDJAGTrustedIssuer

	Resources     []string
	UsernameClaim string
	AMR           *expression.AuthenticationMethodsReferencesMapping
}

// IDJAGClientRelationship is the relationship of a client with a Resource Authorization Server.
type IDJAGClientRelationship struct {
	ClientID                  string
	Scopes                    []string
	Resources                 []string
	AuthorizationDetailsTypes []string
}

// GetIDJAGRelationship returns the relationship of the requesting client with the Resource Authorization Server the
// audience identifies. It returns oauthelia2.ErrNotFound when the audience identifies no Resource Authorization Server
// and oauthelia2.ErrAccessDenied when the client has no relationship with it.
func (s *Store) GetIDJAGRelationship(ctx context.Context, request oauthelia2.AccessRequester, audience string) (relationship *oauthelia2.IDJAGRelationship, err error) {
	ras, ok := s.idjag().LookupResourceAuthorizationServer(audience)
	if !ok {
		return nil, oauthelia2.ErrNotFound.WithDebugf("The audience '%s' does not identify a configured Resource Authorization Server.", audience)
	}

	client, ok := request.GetClient().(Client)
	if !ok {
		return nil, oauthelia2.ErrServerError.WithDebug("The client is not an OpenID Connect 1.0 client.")
	}

	var rel *IDJAGClientRelationship

	if rel, ok = client.GetIDJAGRelationship(ras.Issuer); !ok {
		return nil, oauthelia2.ErrAccessDenied.WithHintf("The client is not permitted to request an Identity Assertion JWT Authorization Grant for the audience '%s'.", audience)
	}

	types := rel.AuthorizationDetailsTypes

	if types == nil {
		types = []string{}
	}

	return &oauthelia2.IDJAGRelationship{
		Issuer:                    ras.Issuer,
		ClientID:                  rel.ClientID,
		Scopes:                    rel.Scopes,
		Resources:                 rel.Resources,
		AuthorizationDetailsTypes: types,
		SigningAlg:                ras.SigningAlg,
		MaxAge:                    ras.MaxAge,
	}, nil
}

// GetIDJAGSubject returns the subject identifier an ID Token issued to the subject client of the Resource
// Authorization Server of the relationship would carry for the End-User of the session.
func (s *Store) GetIDJAGSubject(ctx context.Context, request oauthelia2.AccessRequester, relationship *oauthelia2.IDJAGRelationship) (subject string, err error) {
	var (
		username string
		ras      *IDJAGResourceAuthorizationServer
	)

	if username, ras, err = s.getIDJAGUsernameAndServer(request, relationship); err != nil {
		return "", err
	}

	var client Client

	if client, err = s.GetRegisteredClient(ctx, ras.SubjectClientID); err != nil {
		return "", oauthelia2.ErrServerError.WithWrap(err).WithDebugf("The subject client '%s' of the Resource Authorization Server '%s' could not be loaded.", ras.SubjectClientID, ras.Issuer)
	}

	var opaque uuid.UUID

	if opaque, err = s.GetSubject(ctx, client.GetSectorIdentifierURI(), username); err != nil {
		return "", oauthelia2.ErrServerError.WithWrap(err).WithDebugf("The subject identifier for the Resource Authorization Server '%s' could not be determined.", ras.Issuer)
	}

	return opaque.String(), nil
}

// GetIDJAGClaims returns the claims of the claims policy of the Resource Authorization Server of the relationship for
// the End-User of the session, or nil when it has no claims policy.
func (s *Store) GetIDJAGClaims(ctx context.Context, request oauthelia2.AccessRequester, relationship *oauthelia2.IDJAGRelationship) (claims map[string]any, err error) {
	octx, ok := idjagContext(ctx)
	if !ok {
		return nil, oauthelia2.ErrServerError.WithDebug("The context is not an OpenID Connect 1.0 context.")
	}

	var (
		username string
		ras      *IDJAGResourceAuthorizationServer
	)

	if username, ras, err = s.getIDJAGUsernameAndServer(request, relationship); err != nil {
		return nil, err
	}

	if ras.claims == nil {
		return nil, nil
	}

	resolver := octx.GetProviderUserAttributeResolver()
	if resolver == nil {
		return nil, oauthelia2.ErrServerError.WithDebug("Error occurred obtaining the attribute resolver.")
	}

	users := octx.GetUserProvider()
	if users == nil {
		return nil, oauthelia2.ErrServerError.WithDebug("Error occurred obtaining the user provider.")
	}

	var details *authentication.UserDetailsExtended

	if details, err = getUserDetailsExtended(octx, users, username); err != nil {
		return nil, oauthelia2.ErrServerError.WithWrap(err).WithDebugf("Error occurred retrieving the details of the user '%s'.", username)
	}

	claims = map[string]any{}

	ras.claims.hydrateIDJAGClaims(newResolveFunc(resolver, details, octx.GetClock().Now()), claims)

	delete(claims, ClaimTenant)
	delete(claims, ClaimAudienceTenant)

	return claims, nil
}

func (s *Store) getIDJAGUsernameAndServer(request oauthelia2.AccessRequester, relationship *oauthelia2.IDJAGRelationship) (username string, ras *IDJAGResourceAuthorizationServer, err error) {
	session, ok := request.GetSession().(*Session)
	if !ok || session == nil || session.DefaultSession == nil || session.Username == "" {
		return "", nil, oauthelia2.ErrNotFound.WithDebug("The session does not identify a user.")
	}

	if relationship == nil {
		return "", nil, oauthelia2.ErrServerError.WithDebug("The relationship is not set.")
	}

	idjag := s.idjag()
	if idjag == nil {
		return "", nil, oauthelia2.ErrServerError.WithDebug("The Identity Assertion JWT Authorization Grant is not configured.")
	}

	if ras, ok = idjag.ResourceAuthorizationServers[relationship.Issuer]; !ok {
		return "", nil, oauthelia2.ErrServerError.WithDebugf("The Resource Authorization Server '%s' of the relationship is not configured.", relationship.Issuer)
	}

	return session.Username, ras, nil
}

// GetIDJAGTrustedIssuer returns the trusted issuer with the given issuer identifier, or oauthelia2.ErrNotFound when
// the issuer is not trusted.
func (s *Store) GetIDJAGTrustedIssuer(ctx context.Context, issuer string) (trusted *oauthelia2.IDJAGTrustedIssuer, err error) {
	ti, ok := s.idjag().LookupTrustedIssuer(issuer)
	if !ok {
		return nil, oauthelia2.ErrNotFound.WithDebugf("The issuer '%s' is not a trusted issuer.", issuer)
	}

	return &ti.IDJAGTrustedIssuer, nil
}

// IsIDJAGTenantTrusted returns false as no trusted issuer is multi-tenant.
func (s *Store) IsIDJAGTenantTrusted(ctx context.Context, claims map[string]any) (trusted bool, err error) {
	return false, nil
}

// IsIDJAGResourceGoverned returns true when the resource is one of the resources of the trusted issuer of the grant.
func (s *Store) IsIDJAGResourceGoverned(ctx context.Context, claims map[string]any, resource string) (governed bool, err error) {
	issuer, _ := claims[ClaimIssuer].(string)

	trusted, ok := s.idjag().LookupTrustedIssuer(issuer)
	if !ok {
		return false, nil
	}

	return slices.Contains(trusted.Resources, resource), nil
}

// ResolveIDJAGSubject returns the subject identifier of the user the username claim of the trusted issuer of the grant
// identifies for the sector of the redeeming client. It returns oauthelia2.ErrNotFound when the claim is not a
// non-empty string or the user does not exist. Any other user lookup error is returned unchanged.
func (s *Store) ResolveIDJAGSubject(ctx context.Context, client oauthelia2.Client, claims map[string]any) (subject string, err error) {
	octx, ok := idjagContext(ctx)
	if !ok {
		return "", oauthelia2.ErrServerError.WithDebug("The context is not an OpenID Connect 1.0 context.")
	}

	c, ok := client.(Client)
	if !ok {
		return "", oauthelia2.ErrServerError.WithDebug("The client is not an OpenID Connect 1.0 client.")
	}

	issuer, _ := claims[ClaimIssuer].(string)

	trusted, ok := s.idjag().LookupTrustedIssuer(issuer)
	if !ok {
		return "", oauthelia2.ErrNotFound.WithDebugf("The issuer '%s' is not a trusted issuer.", issuer)
	}

	username, ok := claims[trusted.UsernameClaim].(string)
	if !ok || username == "" {
		return "", oauthelia2.ErrNotFound.WithDebugf("The '%s' claim of the grant from the issuer '%s' is not a non-empty string.", trusted.UsernameClaim, issuer)
	}

	users := octx.GetUserProvider()
	if users == nil {
		return "", oauthelia2.ErrServerError.WithDebug("Error occurred obtaining the user provider.")
	}

	var details *authentication.UserDetailsExtended

	if details, err = getUserDetailsExtended(octx, users, username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			return "", oauthelia2.ErrNotFound.WithWrap(err).WithDebugf("The user '%s' identified by the grant from the issuer '%s' was not found.", username, issuer)
		}

		return "", err
	}

	var opaque uuid.UUID

	if details == nil || details.UserDetails == nil || details.Username == "" {
		return "", oauthelia2.ErrServerError.WithDebugf("The user '%s' identified by the grant from the issuer '%s' has no username.", username, issuer)
	}

	if opaque, err = s.GetSubject(ctx, c.GetSectorIdentifierURI(), details.Username); err != nil {
		return "", err
	}

	return opaque.String(), nil
}

// IsIDJAGUsed returns true when the 'jti' of a grant from the issuer is recorded as used and has not expired.
func (s *Store) IsIDJAGUsed(ctx context.Context, issuer, jti string) (used bool, err error) {
	return isBlacklistedJTIKnown(s.provider.LoadOAuth2BlacklistedIssuerJTI(ctx, issuer, model.NewOAuth2BlacklistedJTISignature(jti)))
}

// MarkIDJAGUsed records the 'jti' of a grant from the issuer as used until exp. It returns oauthelia2.ErrJTIKnown when
// it is already recorded and has not expired.
func (s *Store) MarkIDJAGUsed(ctx context.Context, issuer, jti string, exp time.Time) (err error) {
	return blacklistedJTISaveError(s.provider.SaveOAuth2BlacklistedIssuerJTI(ctx, model.NewOAuth2BlacklistedIssuerJTI(issuer, jti, exp)))
}

func (s *CustomClaimsStrategy) hydrateIDJAGClaims(resolve ClaimResolver, extra map[string]any) {
	for _, claims := range s.scopes {
		for claim, attribute := range claims {
			s.hydrateClaim(nil, claim, attribute, s.claimsIDToken, resolve, extra, nil)
		}
	}
}

func idjagContext(ctx context.Context) (octx Context, ok bool) {
	if octx, ok = ctx.Value(model.CtxKeyAutheliaCtx).(Context); ok {
		return octx, true
	}

	octx, ok = ctx.(Context)

	return octx, ok
}

func sortedUnique(values []string) []string {
	slices.Sort(values)

	return slices.Compact(values)
}

const idjagLifespanDefault = time.Minute * 5

var (
	_ idjag.IssueStorage  = (*Store)(nil)
	_ idjag.RedeemStorage = (*Store)(nil)
)
