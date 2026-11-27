// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Authz returns the restricted view of this strategy for the authorization endpoints.
func (p *DefaultStrategy) Authz() (strategy AuthzStrategy) {
	return p.authz
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
// no such session, and records the request as activity on the session. A request which isn't to an authorization
// endpoint always receives an anonymous session.
func (a *DefaultAuthzStrategy) Get(ctx Context) (session *UserSession, err error) {
	if !ctx.IsAuthzEndpointRequest() {
		userSession := a.base.NewDefault()

		return &userSession, nil
	}

	if cached, ok := a.base.getCached(ctx, a.base.cacheKeyAuthz()); ok {
		return cached, nil
	}

	var (
		sid    string
		record Record
	)

	if sid, _, session, record, err = a.resolveRecord(ctx); err != nil {
		return nil, err
	}

	if session != nil {
		a.base.touch(ctx, sid, record, session)
	} else {
		userSession := a.base.NewDefault()

		session = &userSession
	}

	a.base.setCached(ctx, a.base.cacheKeyAuthz(), session)

	return session, nil
}

// SaveActivity persists the refresh time of the given session onto the stored session, which also records the request
// as activity on it. The expiration of the session is left as it is, as only a save of the session by the portal extends
// it, and the authz cookie isn't reissued as it outlives the session regardless. It does nothing on a request which
// isn't to an authorization endpoint, or when the stored session is anonymous, as the pointer record of an anonymous
// session must never be refreshed.
func (a *DefaultAuthzStrategy) SaveActivity(ctx Context, userSession *UserSession) (err error) {
	if userSession == nil {
		return fmt.Errorf("error occurred saving session activity: it is nil")
	}

	if !ctx.IsAuthzEndpointRequest() {
		return nil
	}

	var (
		sid, aid string
		stored   *UserSession
		record   Record
		data     []byte
	)

	if sid, aid, stored, record, err = a.resolveRecord(ctx); err != nil {
		return err
	}

	p := a.base

	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	if stored == nil || stored.IsAnonymous() {
		return nil
	}

	expiration, ok := p.getRemaining(*stored)
	if !ok {
		return nil
	}

	stored.RefreshTTL = userSession.RefreshTTL

	if data, err = p.codec.Seal(p.config.Domain, sid, *stored); err != nil {
		return fmt.Errorf("error occurred encoding session: %w", err)
	}

	if err = p.repository.SaveData(ctx, p.issuer, sid, stored.PublicID, stored.Username, p.getRetention(*stored, expiration), record.GetSessionData(), data); err != nil {
		return a.handleSaveFailure(ctx, err, "error occurred saving session activity")
	}

	if err = p.repository.SaveData(ctx, p.authzIssuer, aid, getAuthzIndexPublicID(aid), "", expiration, nil, []byte(sid)); err != nil {
		return a.handleSaveFailure(ctx, err, "error occurred saving session authz index")
	}

	p.setCached(ctx, p.cacheKey(), nil)

	return nil
}

// Destroy removes the session the authz cookie of the request is linked to along with the pointer record, and instructs
// the user agent to discard its session cookies. A session which the authz cookie doesn't resolve to is never removed. It
// returns ErrNotPermittedOnPortalRequest on a request which isn't to an authorization endpoint.
func (a *DefaultAuthzStrategy) Destroy(ctx Context) (err error) {
	if !ctx.IsAuthzEndpointRequest() {
		return fmt.Errorf("error occurred destroying session: %w", ErrNotPermittedOnPortalRequest)
	}

	p := a.base

	err = a.remove(ctx)

	p.setCached(ctx, p.cacheKeyAuthz(), nil)
	p.setCached(ctx, p.cacheKey(), nil)

	p.clearCookies(ctx)

	return err
}

func (a *DefaultAuthzStrategy) remove(ctx Context) (err error) {
	p := a.base

	sid, aid, session, err := a.resolve(ctx)
	if err != nil {
		return err
	}

	if session != nil {
		if err = p.repository.Delete(ctx, p.issuer, sid, session.PublicID, session.Username); err != nil {
			return fmt.Errorf("error occurred deleting session from backend: %w", err)
		}
	}

	if aid != "" {
		_ = p.deleteAuthzIndex(ctx, aid)
	}

	return nil
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

func (a *DefaultAuthzStrategy) resolveRecord(ctx Context) (sid, aid string, session *UserSession, record Record, err error) {
	p := a.base

	value := p.getCookieID(ctx, p.authzName)

	if len(value) == 0 {
		return "", "", nil, nil, nil
	}

	aid = p.codec.SignAuthz(value)

	if record, err = p.repository.Get(ctx, p.authzIssuer, aid); err != nil {
		return "", aid, nil, nil, fmt.Errorf("%w: %w", ErrRepositoryGet, err)
	}

	if record == nil || len(record.GetSessionData()) == 0 || string(record.GetSessionData()) == authzIndexMoved {
		return "", aid, nil, nil, nil
	}

	sid = string(record.GetSessionData())

	if record, err = p.repository.Get(ctx, p.issuer, sid); err != nil {
		return "", aid, nil, nil, fmt.Errorf("%w: %w", ErrRepositoryGet, err)
	}

	if record == nil || len(record.GetSessionData()) == 0 || record.GetSessionSignature() != sid {
		logContext(ctx).Debug("Session authz cookie doesn't resolve to a session, which has either expired, exceeded the inactivity timeout, or been destroyed")

		return "", aid, nil, nil, nil
	}

	session = &UserSession{}

	if err = p.codec.Open(p.config.Domain, record, session); err != nil {
		return "", aid, nil, nil, nil
	}

	if _, expected := p.getAuthzValue(sid); session.CookieDomain != p.config.Domain || !equalSignature(expected, aid) {
		return "", aid, nil, nil, nil
	}

	if unanchored, err := p.destroyUnanchoredAuthz(ctx, sid, session); unanchored || err != nil {
		return "", aid, nil, nil, err
	}

	return sid, aid, session, record, nil
}

func (p *DefaultStrategy) destroyUnanchoredAuthz(ctx Context, sid string, session *UserSession) (unanchored bool, err error) {
	if p.config.AnchorRemoteIP == nil || isAnchored(session, ctx.RemoteIP()) {
		return false, nil
	}

	logContext(ctx).Debug("Session authz cookie resolves to a network which isn't anchored to the remote IP")

	if err = p.destroyByID(ctx, sid, session); err != nil {
		return true, fmt.Errorf("error occurred destroying session not anchored to the network of the remote IP: %w", err)
	}

	return true, nil
}

var (
	_ AuthzStrategy = (*DefaultAuthzStrategy)(nil)
)

func (p *DefaultStrategy) retireAuthzIndex(ctx Context, oldSID string) {
	_, aid := p.getAuthzValue(oldSID)

	if err := p.repository.Save(ctx, p.authzIssuer, aid, getAuthzIndexPublicID(aid), "", authzIndexMovedExpiration, []byte(authzIndexMoved)); err != nil {
		_ = p.deleteAuthzIndex(ctx, aid)
	}
}

func (p *DefaultStrategy) healAuthz(ctx Context, sid string, session *UserSession) {
	value, aid := p.getAuthzValue(sid)

	if current := p.getCookieID(ctx, p.authzName); len(current) != 0 && subtle.ConstantTimeCompare(current, value) == 1 {
		return
	}

	remaining, ok := p.getRemaining(*session)
	if !ok {
		return
	}

	if err := p.saveAuthzIndex(ctx, aid, sid, remaining); err != nil {
		return
	}

	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	ctx.SetCookie(p.newAuthzCookie(value, p.getCookieExpires(p.getExpiration(*session))))
}

func (p *DefaultStrategy) clearStaleAuthzCookie(ctx Context, sid string) {
	if ctx.GetCookie(p.authzName) == "" {
		return
	}

	if value := p.getCookieID(ctx, p.authzName); len(value) != 0 {
		record, err := p.repository.Get(ctx, p.authzIssuer, p.codec.SignAuthz(value))
		if err != nil {
			return
		}

		if record != nil && len(record.GetSessionData()) != 0 && string(record.GetSessionData()) != sid {
			return
		}
	}

	p.clearAuthzCookie(ctx)
}

func (p *DefaultStrategy) clearAuthzCookie(ctx Context) {
	if ctx.GetCookie(p.authzName) == "" {
		return
	}

	p.setCached(ctx, p.cacheKeyAuthz(), nil)

	ctx.ClearCookie(p.newAuthzDeletionCookie())
}

func (p *DefaultStrategy) getAuthzValue(sid string) (value []byte, aid string) {
	value, _ = hex.DecodeString(p.codec.SignAuthz([]byte(cookieValuePrefixAuthz + p.issuer + ":" + sid)))

	return value, p.codec.SignAuthz(value)
}

func (p *DefaultStrategy) saveAuthzIndex(ctx Context, aid, sid string, expiration time.Duration) (err error) {
	switch err = p.repository.Save(ctx, p.authzIssuer, aid, getAuthzIndexPublicID(aid), "", expiration, []byte(sid)); {
	case err == nil:
	case errors.Is(err, ErrSessionSuperseded):
		logContext(ctx).WithError(err).Debug("Session authz index was not saved as the session has been destroyed or moved by another request, the authz cookie will not be issued")
	default:
		logContext(ctx).WithError(err).Error("Error occurred saving the session authz index, the authz cookie will not be issued")
	}

	return err
}

func (p *DefaultStrategy) deleteAuthzIndex(ctx Context, aid string) (err error) {
	if err = p.repository.Delete(ctx, p.authzIssuer, aid, getAuthzIndexPublicID(aid), ""); err != nil {
		logContext(ctx).WithError(err).Warn("Error occurred deleting the session authz index")
	}

	return err
}

func equalSignature(a, b string) bool {
	return len(a) != 0 && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func getAuthzIndexPublicID(aid string) string {
	if len(aid) <= authzIndexPublicIDLength {
		return aid
	}

	return aid[:authzIndexPublicIDLength]
}
