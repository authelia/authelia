// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Authz returns the restricted view of this strategy for the authorization endpoints.
func (p *DefaultStrategy) Authz() (strategy AuthzStrategy) {
	return &DefaultAuthzStrategy{base: p}
}

// DefaultAuthzStrategy is the default AuthzStrategy, which follows the pointer record stored against the authz cookie
// to the session record and requires the session to record that authz cookie.
type DefaultAuthzStrategy struct {
	base *DefaultStrategy
}

// GetConfig returns the session cookie configuration of the strategy this view belongs to.
func (a *DefaultAuthzStrategy) GetConfig() (config schema.SessionCookie) {
	return a.base.config
}

// NewDefault returns an anonymous session bound to the cookie domain of this view.
func (a *DefaultAuthzStrategy) NewDefault() (userSession UserSession) {
	return a.base.NewDefault()
}

// Get returns the session the authz cookie of the request is linked to, returning an anonymous session when there is
// no such session. A request which isn't to an authorization endpoint always receives an anonymous session.
func (a *DefaultAuthzStrategy) Get(ctx Context) (session *UserSession, err error) {
	if !ctx.IsAuthzEndpointRequest() {
		userSession := a.base.NewDefault()

		return &userSession, nil
	}

	if cached, ok := a.base.getCached(ctx, a.base.cacheKeyAuthz()); ok {
		return cached, nil
	}

	if _, _, session, err = a.resolve(ctx); err != nil {
		return nil, err
	}

	if session == nil {
		userSession := a.base.NewDefault()

		session = &userSession
	}

	a.base.setCached(ctx, a.base.cacheKeyAuthz(), session)

	return session, nil
}

// SaveActivity persists the last activity and refresh time of the given session onto the stored session. The expiry of
// the session and its pointer record is left as it is, as only a save of the session by the portal extends it, and the
// authz cookie isn't reissued as it outlives the session regardless. It does nothing on a request which isn't to an
// authorization endpoint.
func (a *DefaultAuthzStrategy) SaveActivity(ctx Context, userSession *UserSession) (err error) {
	if userSession == nil {
		return fmt.Errorf("error occurred saving session activity: it is nil")
	}

	if !ctx.IsAuthzEndpointRequest() {
		return nil
	}

	var (
		sid, aid       string
		stored         *UserSession
		previous, data []byte
	)

	if sid, aid, stored, previous, err = a.resolveRecord(ctx); err != nil {
		return err
	}

	p := a.base

	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	if stored == nil {
		return nil
	}

	stored.LastActivity = userSession.LastActivity
	stored.RefreshTTL = userSession.RefreshTTL

	if data, err = p.codec.Seal(p.config.Domain, sid, *stored); err != nil {
		return fmt.Errorf("error occurred encoding session: %w", err)
	}

	expiration, ok := p.getRemaining(*stored)
	if !ok {
		return nil
	}

	if err = p.repository.SaveData(ctx, p.issuer, sid, stored.PublicID, stored.Username, expiration, previous, data); err != nil {
		return a.handleSaveFailure(ctx, err, "error occurred saving session activity")
	}

	if err = p.repository.SaveData(ctx, p.authzIssuer, aid, getAuthzIndexPublicID(aid), "", expiration, nil, []byte(sid)); err != nil {
		return a.handleSaveFailure(ctx, err, "error occurred saving session authz index")
	}

	p.setCached(ctx, p.cacheKey(), nil)

	return nil
}

// Destroy removes the session the authz cookie of the request is linked to along with the pointer record, and instructs
// the user agent to discard both cookies. A session which the authz cookie doesn't resolve to is never removed. It
// returns ErrNotPermittedOnPortalRequest on a request which isn't to an authorization endpoint.
func (a *DefaultAuthzStrategy) Destroy(ctx Context) (err error) {
	if !ctx.IsAuthzEndpointRequest() {
		return fmt.Errorf("error occurred destroying session: %w", ErrNotPermittedOnPortalRequest)
	}

	p := a.base

	sid, aid, session, err := a.resolve(ctx)

	if err == nil && session != nil {
		if err = p.repository.Delete(ctx, p.issuer, sid, session.PublicID, session.Username); err != nil {
			err = fmt.Errorf("error occurred deleting session from backend: %w", err)
		}
	}

	if err == nil && aid != "" {
		_ = p.deleteAuthzIndex(ctx, aid)
	}

	p.setCached(ctx, p.cacheKeyAuthz(), nil)
	p.setCached(ctx, p.cacheKey(), nil)

	ctx.ClearCookie(p.newAuthzDeletionCookie())
	ctx.ClearCookie(p.newDeletionCookie())

	return err
}

func (a *DefaultAuthzStrategy) handleSaveFailure(ctx Context, err error, message string) error {
	if errors.Is(err, ErrSessionSuperseded) {
		return nil
	}

	if _, _, session, rerr := a.resolve(ctx); rerr == nil && session == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", message, err)
}

func (a *DefaultAuthzStrategy) resolve(ctx Context) (sid, aid string, session *UserSession, err error) {
	sid, aid, session, _, err = a.resolveRecord(ctx)

	return sid, aid, session, err
}

// resolveRecord resolves the session the authz cookie of the request is linked to, also returning the sealed data of
// the session record it was opened from.
func (a *DefaultAuthzStrategy) resolveRecord(ctx Context) (sid, aid string, session *UserSession, data []byte, err error) {
	p := a.base

	value := p.getCookieID(ctx, p.authzName)

	if len(value) == 0 {
		return "", "", nil, nil, nil
	}

	aid = p.codec.SignAuthz(value)

	var record Record

	if record, err = p.repository.Get(ctx, p.authzIssuer, aid); err != nil {
		return "", aid, nil, nil, fmt.Errorf("%w: %w", ErrRepositoryGet, err)
	}

	if record == nil || len(record.GetSessionData()) == 0 {
		return "", aid, nil, nil, nil
	}

	sid = string(record.GetSessionData())

	if record, err = p.repository.Get(ctx, p.issuer, sid); err != nil {
		return "", aid, nil, nil, fmt.Errorf("%w: %w", ErrRepositoryGet, err)
	}

	if record == nil || len(record.GetSessionData()) == 0 || record.GetSessionSignature() != sid {
		return "", aid, nil, nil, nil
	}

	session = &UserSession{}

	if err = p.codec.Open(p.config.Domain, record, session); err != nil {
		return "", aid, nil, nil, nil
	}

	if _, expected := p.getAuthzValue(sid); session.CookieDomain != p.config.Domain || !equalSignature(expected, aid) {
		return "", aid, nil, nil, nil
	}

	return sid, aid, session, record.GetSessionData(), nil
}

var (
	_ AuthzStrategy = (*DefaultAuthzStrategy)(nil)
)
