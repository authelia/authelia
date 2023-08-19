package oidc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/token/jwt"
	"github.com/ory/x/errorsx"
	"github.com/pkg/errors"
)

// JWTCoreStrategy wraps the HMACCoreStrategy for the purpose of
// implementing RFC9068 JWT Profile for OAuth 2.0 Access Tokens.
type JWTCoreStrategy struct {
	jwt.Signer
	HMACCoreStrategy *HMACCoreStrategy
	Config           interface {
		fosite.AccessTokenIssuerProvider
		fosite.JWTScopeFieldProvider
	}
}

// AccessTokenSignature implements oauth2.AccessTokenStrategy.
func (s *JWTCoreStrategy) AccessTokenSignature(ctx context.Context, token string) (signature string) {
	if isJWT(token) {
		return s.jwtSignature(token)
	}

	return s.HMACCoreStrategy.AccessTokenSignature(ctx, token)
}

// GenerateAccessToken implements oauth2.AccessTokenStrategy.
func (s *JWTCoreStrategy) GenerateAccessToken(ctx context.Context, requester fosite.Requester) (token string, signature string, err error) {
	var (
		client Client
		ok     bool
	)

	if client, ok = requester.GetClient().(Client); ok && client.GetJWTProfileOAuthAccessTokensEnabled() {
		return s.jwtGenerate(ctx, fosite.AccessToken, requester)
	}

	return s.HMACCoreStrategy.GenerateAccessToken(ctx, requester)
}

// ValidateAccessToken implements oauth2.AccessTokenStrategy.
func (s *JWTCoreStrategy) ValidateAccessToken(ctx context.Context, requester fosite.Requester, token string) (err error) {
	if isJWT(token) {
		_, err = jwtValidate(ctx, s.Signer, token)

		return err
	}

	return s.HMACCoreStrategy.ValidateAccessToken(ctx, requester, token)
}

// RefreshTokenSignature implements oauth2.RefreshTokenStrategy.
func (s *JWTCoreStrategy) RefreshTokenSignature(ctx context.Context, token string) (signature string) {
	return s.HMACCoreStrategy.RefreshTokenSignature(ctx, token)
}

// GenerateRefreshToken implements oauth2.RefreshTokenStrategy.
func (s *JWTCoreStrategy) GenerateRefreshToken(ctx context.Context, req fosite.Requester) (token string, signature string, err error) {
	return s.HMACCoreStrategy.GenerateRefreshToken(ctx, req)
}

// ValidateRefreshToken implements oauth2.RefreshTokenStrategy.
func (s *JWTCoreStrategy) ValidateRefreshToken(ctx context.Context, req fosite.Requester, token string) error {
	return s.HMACCoreStrategy.ValidateRefreshToken(ctx, req, token)
}

// AuthorizeCodeSignature implements oauth2.AuthorizeCodeStrategy.
func (s *JWTCoreStrategy) AuthorizeCodeSignature(ctx context.Context, token string) (signature string) {
	return s.HMACCoreStrategy.AuthorizeCodeSignature(ctx, token)
}

// GenerateAuthorizeCode implements oauth2.AuthorizeCodeStrategy.
func (s *JWTCoreStrategy) GenerateAuthorizeCode(ctx context.Context, req fosite.Requester) (token string, signature string, err error) {
	return s.HMACCoreStrategy.GenerateAuthorizeCode(ctx, req)
}

// ValidateAuthorizeCode implements oauth2.AuthorizeCodeStrategy.
func (s *JWTCoreStrategy) ValidateAuthorizeCode(ctx context.Context, req fosite.Requester, token string) error {
	return s.HMACCoreStrategy.ValidateAuthorizeCode(ctx, req, token)
}

func (s *JWTCoreStrategy) jwtSignature(token string) (signature string) {
	return strings.Split(token, ".")[2]
}

func jwtValidate(ctx context.Context, signer jwt.Signer, rawToken string) (token *jwt.Token, err error) {
	if token, err = signer.Decode(ctx, rawToken); err == nil {
		return token, token.Claims.Valid()
	}

	var e *jwt.ValidationError

	if err != nil && errors.As(err, &e) {
		return token, errorsx.WithStack(jwtValidationErrorToRFC6749Error(e).WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	return token, nil
}

func (s *JWTCoreStrategy) jwtGenerate(ctx context.Context, tokenType fosite.TokenType, requester fosite.Requester) (string, string, error) {
	if jwtSession, ok := requester.GetSession().(oauth2.JWTSessionContainer); !ok {
		return "", "", errors.Errorf("Session must be of type JWTSessionContainer but got type: %T", requester.GetSession())
	} else if jwtSession.GetJWTClaims() == nil {
		return "", "", errors.New("GetTokenClaims() must not be nil")
	} else {
		claims := jwtSession.GetJWTClaims().
			With(
				jwtSession.GetExpiresAt(tokenType),
				requester.GetGrantedScopes(),
				requester.GetGrantedAudience(),
			).
			WithDefaults(
				time.Now().UTC(),
				s.Config.GetAccessTokenIssuer(ctx),
			).
			WithScopeField(
				s.Config.GetJWTScopeField(ctx),
			)

		return s.Signer.Generate(ctx, claims.ToMapClaims(), jwtSession.GetJWTHeader())
	}
}

func isJWT(token string) bool {
	if strings.Count(token, ".") != 2 {
		return false
	}

	return !strings.HasPrefix(token, fmt.Sprintf(tokenPrefixOrgAutheliaFmt, TokenPrefixPartAccessToken))
}

func jwtValidationErrorToRFC6749Error(v *jwt.ValidationError) *fosite.RFC6749Error {
	switch {
	case v == nil:
		return nil
	case v.Has(jwt.ValidationErrorMalformed):
		return fosite.ErrInvalidTokenFormat
	case v.Has(jwt.ValidationErrorUnverifiable | jwt.ValidationErrorSignatureInvalid):
		return fosite.ErrTokenSignatureMismatch
	case v.Has(jwt.ValidationErrorExpired):
		return fosite.ErrTokenExpired
	case v.Has(jwt.ValidationErrorAudience |
		jwt.ValidationErrorIssuedAt |
		jwt.ValidationErrorIssuer |
		jwt.ValidationErrorNotValidYet |
		jwt.ValidationErrorId |
		jwt.ValidationErrorClaimsInvalid):
		return fosite.ErrTokenClaim
	default:
		return fosite.ErrRequestUnauthorized
	}
}
