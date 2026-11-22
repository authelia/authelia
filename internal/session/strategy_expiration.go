// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import "time"

func (p *DefaultStrategy) getExpiration(userSession UserSession) (expiration time.Duration) {
	switch {
	case userSession.Username == "" && p.config.AnonymousExpiration > 0:
		return p.config.AnonymousExpiration
	case userSession.KeepMeLoggedIn && !p.config.DisableRememberMe:
		return p.config.RememberMe
	}

	return p.config.Expiration
}

func (p *DefaultStrategy) getExpiresAt(expiration time.Duration) (expires int64) {
	if expiration == 0 {
		return 0
	}

	return p.clock.Now().Add(expiration).Unix()
}

func (p *DefaultStrategy) getRemaining(session UserSession) (remaining time.Duration, ok bool) {
	if session.Expires == 0 {
		return p.getExpiration(session), true
	}

	remaining = time.Unix(session.Expires, 0).Sub(p.clock.Now())

	return remaining, remaining > 0
}

func (p *DefaultStrategy) getInactivity(session UserSession) (inactivity time.Duration, ok bool) {
	if p.config.Inactivity <= 0 || session.IsAnonymous() || (session.KeepMeLoggedIn && !p.config.DisableRememberMe) {
		return 0, false
	}

	return p.config.Inactivity, true
}

func (p *DefaultStrategy) getRetention(session UserSession, remaining time.Duration) (retention time.Duration) {
	inactivity, ok := p.getInactivity(session)

	switch {
	case !ok:
		return remaining
	case remaining <= 0:
		return inactivity
	default:
		return min(inactivity, remaining)
	}
}

func (p *DefaultStrategy) getTouchInterval(inactivity time.Duration) (interval time.Duration) {
	if p.config.Expiration > 0 {
		inactivity = min(inactivity, p.config.Expiration)
	}

	return min(max(inactivity/sessionTouchDivisor, time.Second), inactivity/2)
}

func (p *DefaultStrategy) touch(ctx Context, sid string, record Record, session *UserSession) {
	inactivity, ok := p.getInactivity(*session)
	if !ok || record.GetSessionTTL() <= 0 || (session.Expires == 0 && p.getExpiration(*session) != 0) {
		return
	}

	remaining, ok := p.getRemaining(*session)
	if !ok {
		return
	}

	retention := p.getRetention(*session, remaining)

	if retention-record.GetSessionTTL() < p.getTouchInterval(inactivity) {
		return
	}

	if err := p.repository.Touch(ctx, p.issuer, sid, session.PublicID, session.Username, retention); err != nil {
		logContext(ctx).WithError(err).Error("Error occurred extending the session for activity")
	}
}
