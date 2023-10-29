package oidc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/storage"
	"github.com/ory/x/errorsx"
	"github.com/pkg/errors"
)

// RefreshTokenGrantHandler handles access requests for the Refresh Token Flow.
type RefreshTokenGrantHandler struct {
	AccessTokenStrategy    oauth2.AccessTokenStrategy
	RefreshTokenStrategy   oauth2.RefreshTokenStrategy
	TokenRevocationStorage oauth2.TokenRevocationStorage
	Config                 interface {
		fosite.AccessTokenLifespanProvider
		fosite.RefreshTokenLifespanProvider
		fosite.ScopeStrategyProvider
		fosite.AudienceStrategyProvider
		fosite.RefreshTokenScopesProvider
	}
}

// HandleTokenEndpointRequest implements https://tools.ietf.org/html/rfc6749#section-6
//
//nolint:gocyclo
func (c *RefreshTokenGrantHandler) HandleTokenEndpointRequest(ctx context.Context, request fosite.AccessRequester) error {
	if !c.CanHandleTokenEndpointRequest(ctx, request) {
		return errorsx.WithStack(fosite.ErrUnknownRequest)
	}

	if !request.GetClient().GetGrantTypes().Has(GrantTypeRefreshToken) {
		return errorsx.WithStack(fosite.ErrUnauthorizedClient.WithHint("The OAuth 2.0 Client is not allowed to use authorization grant 'refresh_token'."))
	}

	refresh := request.GetRequestForm().Get(FormParameterRefreshToken)
	signature := c.RefreshTokenStrategy.RefreshTokenSignature(ctx, refresh)
	originalRequest, err := c.TokenRevocationStorage.GetRefreshTokenSession(ctx, signature, request.GetSession())

	switch {
	case err == nil:
		if err = c.RefreshTokenStrategy.ValidateRefreshToken(ctx, originalRequest, refresh); err != nil {
			// The authorization server MUST ... validate the refresh token.
			// This needs to happen after store retrieval for the session to be hydrated properly.
			if errors.Is(err, fosite.ErrTokenExpired) {
				return errorsx.WithStack(fosite.ErrInvalidGrant.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
			}

			return errorsx.WithStack(fosite.ErrInvalidRequest.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
		}
	case errors.Is(err, fosite.ErrInactiveToken):
		// Detected refresh token reuse.
		if e := c.handleRefreshTokenReuse(ctx, signature, originalRequest); e != nil {
			return errorsx.WithStack(e)
		}

		return errorsx.WithStack(fosite.ErrInactiveToken.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	case errors.Is(err, fosite.ErrNotFound):
		return errorsx.WithStack(fosite.ErrInvalidGrant.WithWrap(err).WithDebugf("The refresh token has not been found: %s", ErrorToDebugRFC6749Error(err).Error()))
	default:
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	if !(len(c.Config.GetRefreshTokenScopes(ctx)) == 0 || originalRequest.GetGrantedScopes().HasOneOf(c.Config.GetRefreshTokenScopes(ctx)...)) {
		scopeNames := strings.Join(c.Config.GetRefreshTokenScopes(ctx), " or ")
		hint := fmt.Sprintf("The OAuth 2.0 Client was not granted scope %s and may thus not perform the 'refresh_token' authorization grant.", scopeNames)

		return errorsx.WithStack(fosite.ErrScopeNotGranted.WithHint(hint))
	}

	// The authorization server MUST ... and ensure that the refresh token was issued to the authenticated client.
	if originalRequest.GetClient().GetID() != request.GetClient().GetID() {
		return errorsx.WithStack(fosite.ErrInvalidGrant.WithHint("The OAuth 2.0 Client ID from this request does not match the ID during the initial token issuance."))
	}

	request.SetID(originalRequest.GetID())
	request.SetSession(originalRequest.GetSession().Clone())

	/*
			There are two key points in the following spec section this addresses:
				1. If omitted the scope param should be treated as the same as the scope originally granted by the resource owner.
				2. The REQUESTED scope MUST NOT include any scope not originally granted.

			scope
					OPTIONAL.  The scope of the access request as described by Section 3.3.  The requested scope MUST NOT
		  			include any scope not originally granted by the resource owner, and if omitted is treated as equal to
		   			the scope originally granted by the resource owner.

			See https://datatracker.ietf.org/doc/html/rfc6749#section-6
	*/

	// Addresses point 1 of the text in RFC6749 Section 6.
	if len(request.GetRequestedScopes()) == 0 {
		request.SetRequestedScopes(originalRequest.GetGrantedScopes())
	}

	request.SetRequestedAudience(originalRequest.GetRequestedAudience())

	strategy := c.Config.GetScopeStrategy(ctx)
	originalScopes := originalRequest.GetGrantedScopes()

	for _, scope := range request.GetRequestedScopes() {
		if !originalScopes.Has(scope) {
			if client, ok := request.GetClient().(RefreshFlowScopeClient); ok && client.GetRefreshFlowIgnoreOriginalGrantedScopes(ctx) {
				// Skips addressing point 2 of the text in RFC6749 Section 6 and instead just prevents the scope
				// requested from being granted.
				continue
			}

			// Addresses point 2 of the text in RFC6749 Section 6.
			return errorsx.WithStack(fosite.ErrInvalidScope.WithHintf("The requested scope '%s' was not originally granted by the resource owner.", scope))
		}

		if !strategy(request.GetClient().GetScopes(), scope) {
			return errorsx.WithStack(fosite.ErrInvalidScope.WithHintf("The OAuth 2.0 Client is not allowed to request scope '%s'.", scope))
		}

		request.GrantScope(scope)
	}

	if err = c.Config.GetAudienceStrategy(ctx)(request.GetClient().GetAudience(), originalRequest.GetGrantedAudience()); err != nil {
		return err
	}

	for _, audience := range originalRequest.GetGrantedAudience() {
		request.GrantAudience(audience)
	}

	atLifespan := fosite.GetEffectiveLifespan(request.GetClient(), fosite.GrantTypeRefreshToken, fosite.AccessToken, c.Config.GetAccessTokenLifespan(ctx))
	request.GetSession().SetExpiresAt(fosite.AccessToken, time.Now().UTC().Add(atLifespan).Round(time.Second))

	rtLifespan := fosite.GetEffectiveLifespan(request.GetClient(), fosite.GrantTypeRefreshToken, fosite.RefreshToken, c.Config.GetRefreshTokenLifespan(ctx))
	if rtLifespan > -1 {
		request.GetSession().SetExpiresAt(fosite.RefreshToken, time.Now().UTC().Add(rtLifespan).Round(time.Second))
	}

	return nil
}

// PopulateTokenEndpointResponse implements https://tools.ietf.org/html/rfc6749#section-6
func (c *RefreshTokenGrantHandler) PopulateTokenEndpointResponse(ctx context.Context, requester fosite.AccessRequester, responder fosite.AccessResponder) (err error) {
	if !c.CanHandleTokenEndpointRequest(ctx, requester) {
		return errorsx.WithStack(fosite.ErrUnknownRequest)
	}

	var (
		accessToken, refreshToken         string
		accessSignature, refreshSignature string
	)

	if accessToken, accessSignature, err = c.AccessTokenStrategy.GenerateAccessToken(ctx, requester); err != nil {
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	if refreshToken, refreshSignature, err = c.RefreshTokenStrategy.GenerateRefreshToken(ctx, requester); err != nil {
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	signature := c.RefreshTokenStrategy.RefreshTokenSignature(ctx, requester.GetRequestForm().Get(GrantTypeRefreshToken))

	if ctx, err = storage.MaybeBeginTx(ctx, c.TokenRevocationStorage); err != nil {
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	defer func() {
		err = c.handleRefreshTokenEndpointStorageError(ctx, err)
	}()

	var original fosite.Requester

	if original, err = c.TokenRevocationStorage.GetRefreshTokenSession(ctx, signature, nil); err != nil {
		return err
	}

	if err = c.TokenRevocationStorage.RevokeAccessToken(ctx, original.GetID()); err != nil {
		return err
	}

	if err = c.TokenRevocationStorage.RevokeRefreshTokenMaybeGracePeriod(ctx, original.GetID(), signature); err != nil {
		return err
	}

	if err = c.TokenRevocationStorage.CreateAccessTokenSession(ctx, accessSignature, RefreshFlowSanitizeRestoreOriginalRequestBasic(requester, original)); err != nil {
		return err
	}

	if err = c.TokenRevocationStorage.CreateRefreshTokenSession(ctx, refreshSignature, RefreshFlowSanitizeRestoreOriginalRequest(requester, original)); err != nil {
		return err
	}

	responder.SetAccessToken(accessToken)
	responder.SetTokenType(fosite.BearerAccessToken)
	responder.SetExpiresIn(getExpiresIn(requester, fosite.AccessToken, fosite.GetEffectiveLifespan(requester.GetClient(), fosite.GrantTypeRefreshToken, fosite.AccessToken, c.Config.GetAccessTokenLifespan(ctx)), time.Now().UTC()))
	responder.SetScopes(requester.GetGrantedScopes())
	responder.SetExtra(GrantTypeRefreshToken, refreshToken)

	if err = storage.MaybeCommitTx(ctx, c.TokenRevocationStorage); err != nil {
		return err
	}

	return nil
}

// Reference: https://tools.ietf.org/html/rfc6819#section-5.2.2.3
//
//	The basic idea is to change the refresh token
//	value with every refresh request in order to detect attempts to
//	obtain access tokens using old refresh tokens.  Since the
//	authorization server cannot determine whether the attacker or the
//	legitimate client is trying to access, in case of such an access
//	attempt the valid refresh token and the access authorization
//	associated with it are both revoked.
func (c *RefreshTokenGrantHandler) handleRefreshTokenReuse(ctx context.Context, signature string, req fosite.Requester) (err error) {
	if ctx, err = storage.MaybeBeginTx(ctx, c.TokenRevocationStorage); err != nil {
		return errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	defer func() {
		err = c.handleRefreshTokenEndpointStorageError(ctx, err)
	}()

	if err = c.TokenRevocationStorage.DeleteRefreshTokenSession(ctx, signature); err != nil {
		return err
	}

	if err = c.TokenRevocationStorage.RevokeRefreshToken(ctx, req.GetID()); err != nil && !errors.Is(err, fosite.ErrNotFound) {
		return err
	}

	if err = c.TokenRevocationStorage.RevokeAccessToken(ctx, req.GetID()); err != nil && !errors.Is(err, fosite.ErrNotFound) {
		return err
	}

	if err = storage.MaybeCommitTx(ctx, c.TokenRevocationStorage); err != nil {
		return err
	}

	return nil
}

func (c *RefreshTokenGrantHandler) handleRefreshTokenEndpointStorageError(ctx context.Context, storageErr error) (err error) {
	if storageErr == nil {
		return nil
	}

	defer func() {
		if rollBackTxnErr := storage.MaybeRollbackTx(ctx, c.TokenRevocationStorage); rollBackTxnErr != nil {
			err = errorsx.WithStack(fosite.ErrServerError.WithWrap(err).WithDebugf("error: %s; rollback error: %s", err, rollBackTxnErr))
		}
	}()

	if errors.Is(storageErr, fosite.ErrSerializationFailure) {
		return errorsx.WithStack(fosite.ErrInvalidRequest.
			WithDebug(ErrorToDebugRFC6749Error(storageErr).Error()).
			WithHint("Failed to refresh token because of multiple concurrent requests using the same token which is not allowed."))
	}

	if errors.Is(storageErr, fosite.ErrNotFound) || errors.Is(storageErr, fosite.ErrInactiveToken) {
		return errorsx.WithStack(fosite.ErrInvalidRequest.
			WithDebug(ErrorToDebugRFC6749Error(storageErr).Error()).
			WithHint("Failed to refresh token because of multiple concurrent requests using the same token which is not allowed."))
	}

	return errorsx.WithStack(fosite.ErrServerError.WithWrap(storageErr).WithDebug(ErrorToDebugRFC6749Error(storageErr).Error()))
}

func (c *RefreshTokenGrantHandler) CanSkipClientAuth(ctx context.Context, requester fosite.AccessRequester) bool {
	return false
}

func (c *RefreshTokenGrantHandler) CanHandleTokenEndpointRequest(ctx context.Context, requester fosite.AccessRequester) bool {
	// grant_type REQUIRED.
	// Value MUST be set to "refresh_token".
	return requester.GetGrantTypes().ExactOne(GrantTypeRefreshToken)
}

// RefreshFlowSanitizeRestoreOriginalRequest sanitizes input requester with the ID of original, and if the underlying type
// of requester is a *fosite.AccessRequest it also restores the originally granted scopes. This ensures the granted
// scopes for the refresh token session never change and can be referenced when determining if a session can grant
// the respective scopes.
func RefreshFlowSanitizeRestoreOriginalRequest(requester, original fosite.Requester) fosite.Requester {
	var (
		ar *fosite.AccessRequest
		ok bool
	)

	if ar, ok = requester.(*fosite.AccessRequest); !ok {
		return RefreshFlowSanitizeRestoreOriginalRequestBasic(requester, original)
	}

	var sr *fosite.Request

	if sr, ok = ar.Sanitize(nil).(*fosite.Request); !ok {
		return RefreshFlowSanitizeRestoreOriginalRequestBasic(requester, original)
	}

	sr.SetID(original.GetID())

	sr.SetRequestedScopes(original.GetRequestedScopes())
	sr.GrantedScope = original.GetGrantedScopes()

	return sr
}

// RefreshFlowSanitizeRestoreOriginalRequestBasic is the fallback sanitizer if the RefreshFlowSanitizeRestoreOriginalRequest
// function fails to assert the requester as *fosite.AccessRequest.
func RefreshFlowSanitizeRestoreOriginalRequestBasic(r, o fosite.Requester) fosite.Requester {
	sr := r.Sanitize(nil)
	sr.SetID(o.GetID())

	return sr
}

var (
	_ fosite.TokenEndpointHandler = (*RefreshTokenGrantHandler)(nil)
)
