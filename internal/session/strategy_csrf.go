// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"fmt"
)

// VerifyCSRFToken returns nil if the token is the CSRF token of the established session of the request, comparing them
// in constant time. It returns ErrCSRFTokenNoSession when the request has no established session, ErrCSRFTokenInvalid
// when the token doesn't match, and any other error when the session couldn't be retrieved.
func (p *DefaultStrategy) VerifyCSRFToken(ctx Context, token string) (err error) {
	var session *UserSession

	if _, session, err = p.getCSRFSession(ctx); err != nil {
		return err
	}

	if len(session.CSRF) == 0 || len(token) == 0 || !p.codec.VerifyCSRF(p.csrfTokenMessage(session.CSRF), token) {
		return ErrCSRFTokenInvalid
	}

	return nil
}

// SetCSRFCookie delivers the CSRF token of the established session of the request to the user agent, generating a CSRF
// secret for an established session which has none. A request without an established session has no token to deliver,
// which isn't an error.
func (p *DefaultStrategy) SetCSRFCookie(ctx Context) (err error) {
	var (
		sid       string
		session   *UserSession
		persisted bool
	)

	if sid, session, err = p.getCSRFSession(ctx); err != nil {
		if errors.Is(err, ErrCSRFTokenNoSession) {
			return nil
		}

		return err
	}

	if len(session.CSRF) == 0 {
		if persisted, err = p.establishCSRFSecret(ctx, sid, session); err != nil {
			return err
		}

		if !persisted {
			session.CSRF = nil

			return nil
		}

		p.setCached(ctx, p.cacheKey(), session)
	}

	ctx.SetCookie(p.newCSRFCookie(session.CSRF, p.getCookieExpires(p.getExpiration(*session))))

	return nil
}

func (p *DefaultStrategy) getCSRFSession(ctx Context) (sid string, session *UserSession, err error) {
	if ctx.IsAuthzEndpointRequest() {
		return "", nil, ErrCSRFTokenNoSession
	}

	if sid, session, err = p.get(ctx, false); err != nil {
		return "", nil, err
	}

	if session == nil || session.PublicID == "" {
		return "", nil, ErrCSRFTokenNoSession
	}

	return sid, session, nil
}

func (p *DefaultStrategy) adoptCSRFSecret(ctx Context, session *UserSession) {
	if _, ok := ctx.(CachingContext); !ok {
		session.CSRF = nil

		return
	}

	if cached, ok := p.getCached(ctx, p.cacheKey()); ok && len(cached.CSRF) != 0 && cached.PublicID == session.PublicID {
		session.CSRF = cached.CSRF
	}
}

func (p *DefaultStrategy) ensureCSRFSecret(ctx Context, id []byte, existing bool, session *UserSession) (err error) {
	if len(session.CSRF) != 0 {
		return nil
	}

	if existing {
		_, err = p.establishCSRFSecret(ctx, p.codec.Sign(id), session)

		return err
	}

	return p.generateCSRFSecret(session)
}

func (p *DefaultStrategy) establishCSRFSecret(ctx Context, sid string, session *UserSession) (persisted bool, err error) {
	for range csrfSecretEstablishAttempts {
		var (
			record Record
			data   []byte
		)

		if record, err = p.repository.Get(ctx, p.issuer, sid); err != nil {
			return false, fmt.Errorf("%w: %w", ErrRepositoryGet, err)
		}

		stored := &UserSession{}

		if record == nil || len(record.GetSessionData()) == 0 || record.GetSessionSignature() != sid || p.codec.Open(p.config.Domain, record, stored) != nil || stored.CookieDomain != p.config.Domain {
			return false, p.generateCSRFSecret(session)
		}

		if len(stored.CSRF) != 0 {
			session.CSRF = stored.CSRF

			return true, nil
		}

		remaining, ok := p.getRemaining(*stored)
		if !ok {
			return false, p.generateCSRFSecret(session)
		}

		if err = p.generateCSRFSecret(stored); err != nil {
			return false, err
		}

		if data, err = p.codec.Seal(p.config.Domain, sid, *stored); err != nil {
			return false, fmt.Errorf("error occurred encoding session: %w", err)
		}

		switch err = p.repository.SaveData(ctx, p.issuer, sid, stored.PublicID, stored.Username, p.getRetention(*stored, remaining), record.GetSessionData(), data); {
		case err == nil:
			session.CSRF = stored.CSRF

			return true, nil
		case !errors.Is(err, ErrSessionSuperseded):
			return false, fmt.Errorf("error occurred saving session CSRF secret: %w", err)
		}
	}

	return false, fmt.Errorf("error occurred establishing session CSRF secret: %w", ErrSessionSuperseded)
}

func (p *DefaultStrategy) generateCSRFSecret(session *UserSession) (err error) {
	if session.CSRF, err = p.codec.GenerateCSRFSecret(); err != nil {
		return fmt.Errorf("error occurred generating session CSRF secret: %w", err)
	}

	return nil
}

func (p *DefaultStrategy) csrfToken(secret []byte) (token string) {
	return p.codec.SignCSRF(p.csrfTokenMessage(secret))
}

func (p *DefaultStrategy) csrfTokenMessage(secret []byte) (message []byte) {
	message = make([]byte, 0, len(cookieValuePrefixCSRFToken)+len(p.config.Domain)+1+len(secret))

	message = append(message, cookieValuePrefixCSRFToken...)
	message = append(message, p.config.Domain...)
	message = append(message, ':')

	return append(message, secret...)
}
