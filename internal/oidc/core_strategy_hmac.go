package oidc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/token/hmac"
	"github.com/ory/x/errorsx"
)

// HMACCoreStrategy implements oauth2.CoreStrategy. It's a copy of the oauth2.HMACSHAStrategy.
type HMACCoreStrategy struct {
	Enigma *hmac.HMACStrategy
	Config interface {
		fosite.AccessTokenLifespanProvider
		fosite.RefreshTokenLifespanProvider
		fosite.AuthorizeCodeLifespanProvider
	}
}

// AccessTokenSignature implements oauth2.AccessTokenStrategy.
func (h *HMACCoreStrategy) AccessTokenSignature(ctx context.Context, tokenString string) string {
	return h.Enigma.Signature(tokenString)
}

// GenerateAccessToken implements oauth2.AccessTokenStrategy.
func (h *HMACCoreStrategy) GenerateAccessToken(ctx context.Context, _ fosite.Requester) (tokenString string, sig string, err error) {
	if tokenString, sig, err = h.Enigma.Generate(ctx); err != nil {
		return "", "", err
	}

	return h.setPrefix(tokenString, TokenPrefixPartAccessToken), sig, nil
}

// ValidateAccessToken implements oauth2.AccessTokenStrategy.
func (h *HMACCoreStrategy) ValidateAccessToken(ctx context.Context, r fosite.Requester, tokenString string) (err error) {
	var exp = r.GetSession().GetExpiresAt(fosite.AccessToken)
	if exp.IsZero() && r.GetRequestedAt().Add(h.Config.GetAccessTokenLifespan(ctx)).Before(time.Now().UTC()) {
		return errorsx.WithStack(fosite.ErrTokenExpired.WithHintf("Access token expired at '%s'.", r.GetRequestedAt().Add(h.Config.GetAccessTokenLifespan(ctx))))
	}

	if !exp.IsZero() && exp.Before(time.Now().UTC()) {
		return errorsx.WithStack(fosite.ErrTokenExpired.WithHintf("Access token expired at '%s'.", exp))
	}

	return h.Enigma.Validate(ctx, h.trimPrefix(tokenString, TokenPrefixPartAccessToken))
}

// RefreshTokenSignature implements oauth2.RefreshTokenStrategy.
func (h *HMACCoreStrategy) RefreshTokenSignature(ctx context.Context, tokenString string) string {
	return h.Enigma.Signature(tokenString)
}

// GenerateRefreshToken implements oauth2.RefreshTokenStrategy.
func (h *HMACCoreStrategy) GenerateRefreshToken(ctx context.Context, _ fosite.Requester) (tokenString string, sig string, err error) {
	if tokenString, sig, err = h.Enigma.Generate(ctx); err != nil {
		return "", "", err
	}

	return h.setPrefix(tokenString, TokenPrefixPartRefreshToken), sig, nil
}

// ValidateRefreshToken implements oauth2.RefreshTokenStrategy.
func (h *HMACCoreStrategy) ValidateRefreshToken(ctx context.Context, r fosite.Requester, tokenString string) (err error) {
	var exp = r.GetSession().GetExpiresAt(fosite.RefreshToken)

	if exp.IsZero() {
		return h.Enigma.Validate(ctx, h.trimPrefix(tokenString, TokenPrefixPartRefreshToken))
	}

	if exp.Before(time.Now().UTC()) {
		return errorsx.WithStack(fosite.ErrTokenExpired.WithHintf("Refresh token expired at '%s'.", exp))
	}

	return h.Enigma.Validate(ctx, h.trimPrefix(tokenString, TokenPrefixPartRefreshToken))
}

// AuthorizeCodeSignature implements oauth2.AuthorizeCodeStrategy.
func (h *HMACCoreStrategy) AuthorizeCodeSignature(ctx context.Context, token string) string {
	return h.Enigma.Signature(token)
}

// GenerateAuthorizeCode implements oauth2.AuthorizeCodeStrategy.
func (h *HMACCoreStrategy) GenerateAuthorizeCode(ctx context.Context, _ fosite.Requester) (tokenString string, sig string, err error) {
	if tokenString, sig, err = h.Enigma.Generate(ctx); err != nil {
		return "", "", err
	}

	return h.setPrefix(tokenString, TokenPrefixPartAuthorizeCode), sig, nil
}

// ValidateAuthorizeCode implements oauth2.AuthorizeCodeStrategy.
func (h *HMACCoreStrategy) ValidateAuthorizeCode(ctx context.Context, r fosite.Requester, tokenString string) (err error) {
	var exp = r.GetSession().GetExpiresAt(fosite.AuthorizeCode)

	if exp.IsZero() && r.GetRequestedAt().Add(h.Config.GetAuthorizeCodeLifespan(ctx)).Before(time.Now().UTC()) {
		return errorsx.WithStack(fosite.ErrTokenExpired.WithHintf("Authorize code expired at '%s'.", r.GetRequestedAt().Add(h.Config.GetAuthorizeCodeLifespan(ctx))))
	}

	if !exp.IsZero() && exp.Before(time.Now().UTC()) {
		return errorsx.WithStack(fosite.ErrTokenExpired.WithHintf("Authorize code expired at '%s'.", exp))
	}

	return h.Enigma.Validate(ctx, h.trimPrefix(tokenString, TokenPrefixPartAuthorizeCode))
}

func (h *HMACCoreStrategy) getPrefix(part string) string {
	return h.getCustomPrefix(tokenPrefixOrgAutheliaFmt, part)
}

func (h *HMACCoreStrategy) getCustomPrefix(tokenPrefixFmt, part string) string {
	return fmt.Sprintf(tokenPrefixFmt, part)
}

func (h *HMACCoreStrategy) setPrefix(tokenString, part string) string {
	return h.getPrefix(part) + tokenString
}

func (h *HMACCoreStrategy) trimPrefix(tokenString, part string) string {
	if strings.HasPrefix(tokenString, h.getCustomPrefix(tokenPrefixOrgOryFmt, part)) {
		return strings.TrimPrefix(tokenString, h.getCustomPrefix(tokenPrefixOrgOryFmt, part))
	}

	return strings.TrimPrefix(tokenString, h.getPrefix(part))
}
