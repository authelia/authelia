// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"encoding/base64"
	"net"
	"net/http"
	"strings"
	"time"
)

func (p *DefaultStrategy) clearCookies(ctx Context) {
	ctx.ClearCookie(p.newDeletionCookie())
	ctx.ClearCookie(p.newAuthzDeletionCookie())
	ctx.ClearCookie(p.newCSRFDeletionCookie())
}

func (p *DefaultStrategy) newUserAgentBindingCookie(binding string) (cookie *http.Cookie) {
	return &http.Cookie{
		Name:     p.bindingName,
		Value:    binding,
		Path:     "/",
		MaxAge:   cookieMaxAgeUserAgentBinding,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// SessionCookieName returns the name of the session cookie for the given configured cookie name. The session cookie is
// only honored on the Authelia URL host and is never read by scripts, so it's host only and HTTP only and carries the
// prefix which requires it to be both.
func SessionCookieName(name string) string {
	return CookieNamePrefixHostHTTP + name
}

// UserAgentBindingCookieName returns the name of the user agent binding cookie for the given configured cookie
// name. The cookie is host only and HTTP only so it carries the same prefix as the session cookie, which prevents a
// sibling subdomain or a script from setting a cookie which takes its place.
func UserAgentBindingCookieName(name string) string {
	return CookieNamePrefixHostHTTP + name + CookieNameSuffixUserAgentBinding
}

// AuthzCookieName returns the name of the authz cookie for the given configured cookie name. The authz cookie is scoped
// to the cookie domain, so it carries the prefix which only requires it to be secure.
func AuthzCookieName(name string) string {
	return CookieNamePrefixSecure + name + CookieNameSuffixAuthz
}

// CSRFCookieName returns the name of the cookie which delivers the CSRF token of the session stored in the cookie with
// the given name to the user agent.
func CSRFCookieName(name string) string {
	return CookieNamePrefixHost + name + CookieNameSuffixCSRFToken
}

func (p *DefaultStrategy) newCSRFCookie(secret []byte, expires time.Time) (cookie *http.Cookie) {
	return p.newCSRFCookieValue(p.csrfToken(secret), expires)
}

func (p *DefaultStrategy) newCSRFDeletionCookie() (cookie *http.Cookie) {
	return p.newCSRFCookieValue("", p.clock.Now().Add(-cookieDeletionOffset))
}

func (p *DefaultStrategy) newCSRFCookieValue(value string, expires time.Time) (cookie *http.Cookie) {
	//nolint:gosec // The cookie is deliberately readable by scripts, and carries a token derived from the session rather than the session identifier.
	return &http.Cookie{
		Name:     CSRFCookieName(p.config.Name),
		Value:    value,
		Path:     "/",
		Expires:  expires,
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	}
}

func (p *DefaultStrategy) getCookieID(ctx Context, name string) (id []byte) {
	value := ctx.GetCookie(name)

	if len(value) == 0 {
		return nil
	}

	var err error

	if id, err = base64.RawURLEncoding.DecodeString(value); err != nil || len(id) != sessionIDLength {
		return nil
	}

	return id
}

func (p *DefaultStrategy) newCookie(id []byte, expires time.Time) (cookie *http.Cookie) {
	return &http.Cookie{
		Name:     p.name,
		Value:    encodeCookieID(id),
		Path:     "/",
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func (p *DefaultStrategy) newAuthzCookie(id []byte, expires time.Time) (cookie *http.Cookie) {
	//nolint:gosec // The SameSite attribute is determined by the validated configuration which restricts it to 'none', 'lax', or 'strict'.
	return &http.Cookie{
		Name:     p.authzName,
		Value:    encodeCookieID(id),
		Path:     "/",
		Domain:   p.authzDomain,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: p.authzSameSite,
	}
}

func (p *DefaultStrategy) newAuthzDeletionCookie() (cookie *http.Cookie) {
	return p.newAuthzCookie(nil, p.clock.Now().Add(-cookieDeletionOffset))
}

func (p *DefaultStrategy) newDeletionCookie() (cookie *http.Cookie) {
	return p.newCookie(nil, p.clock.Now().Add(-cookieDeletionOffset))
}

func (p *DefaultStrategy) getCookieExpires(expiration time.Duration) (exp time.Time) {
	if expiration == 0 {
		return expireUnlimited
	}

	lifetime := max(expiration, sessionCookieLifetimeMinimum)

	if !p.config.DisableRememberMe {
		lifetime = max(lifetime, p.config.RememberMe)
	}

	return p.clock.Now().Add(lifetime)
}

func encodeCookieID(id []byte) string {
	return base64.RawURLEncoding.EncodeToString(id)
}

func newDomain(value string, samSite http.SameSite) string {
	switch {
	case strings.HasPrefix(value, "."), samSite == http.SameSiteStrictMode, net.ParseIP(value) != nil:
		return value
	default:
		return "." + value
	}
}

func newSameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
