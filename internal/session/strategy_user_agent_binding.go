// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// UserAgentBinding returns the opaque value which binds a record to the user agent of the request, which is the
// value of the user agent binding cookie of the request when it carries a well formed one, and otherwise a new random
// value. The binding is not issued to the user agent until IssueUserAgentBinding is called with it.
func (p *DefaultStrategy) UserAgentBinding(ctx Context) (binding string, err error) {
	id := p.getCookieID(ctx, p.bindingName)

	if len(id) == 0 {
		if id, err = p.codec.GenerateSessionID(); err != nil {
			return "", fmt.Errorf("error occurred generating user agent binding: %w", err)
		}
	}

	return encodeCookieID(id), nil
}

// IssueUserAgentBinding issues the binding to the user agent as the value of the user agent binding cookie, which
// also extends the lifetime of a binding the user agent already holds.
func (p *DefaultStrategy) IssueUserAgentBinding(ctx Context, binding string) {
	ctx.SetCookie(p.newUserAgentBindingCookie(binding))
}

// VerifyUserAgentBinding returns true when the digest is the UserAgentBindingDigest of the value of the
// user agent binding cookie of the request, comparing them in constant time.
func (p *DefaultStrategy) VerifyUserAgentBinding(ctx Context, digest string) (valid bool) {
	if len(digest) == 0 || len(p.getCookieID(ctx, p.bindingName)) == 0 {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(UserAgentBindingDigest(ctx.GetCookie(p.bindingName))), []byte(digest)) == 1
}

// UserAgentBindingDigest returns the digest of a binding which is stored with the record it binds, so the stored
// value can't be used as a user agent binding cookie.
func UserAgentBindingDigest(binding string) (digest string) {
	sum := sha256.Sum256([]byte(binding))

	return hex.EncodeToString(sum[:])
}
