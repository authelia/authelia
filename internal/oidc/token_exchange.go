// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"github.com/google/uuid"
)

// HydrateTokenExchangeSessionWithRequestingClient re-establishes the identity of the client performing an RFC 8693
// Token Exchange, and carries the identity of the subject, on the session used to issue the exchanged token.
//
// The RFC 8693 token type handlers load the 'subject_token' (and the 'actor_token') into fresh sessions, so the live
// exchange session is a new session populated only by the subject, the authentication context, the subject and actor
// token claims and the subject requester. Anything else the subject token's session held, such as the username and
// whether it was issued via the client credentials grant, must be carried over explicitly from the subject requester
// when the subject token was a stored token.
//
// Without this the exchanged token would have no username, and would be attributed to the wrong client: an ID Token
// minted from an access or refresh token would carry the original client as its 'azp', and would inherit that client's
// audience, nonce, expiry and the 'at_hash', 'c_hash' and 's_hash' bindings of an authorization response it was never
// part of. The 'may_act' claim authorizes actors against the subject token only and is never propagated.
//
// See: https://datatracker.ietf.org/doc/html/rfc8693
func HydrateTokenExchangeSessionWithRequestingClient(client Client, session *Session) {
	if client == nil || session == nil {
		return
	}

	session.ClientID = client.GetID()

	// The consent session of the subject token belongs to the original client and its grant does not extend to the
	// requesting client.
	session.ChallengeID = uuid.NullUUID{}

	if subject := tokenExchangeSubjectSession(session); subject != nil {
		session.ClientCredentials = subject.ClientCredentials

		if subject.DefaultSession != nil && session.DefaultSession != nil {
			session.Username = subject.Username
		}
	}

	delete(session.Extra, ClaimAuthorizedActor)

	if session.DefaultSession == nil || session.Claims == nil {
		return
	}

	session.Claims.AuthorizedParty = client.GetID()
	session.Claims.Audience = nil
	session.Claims.Nonce = ""
	session.Claims.ExpirationTime = nil
	session.Claims.AccessTokenHash = ""
	session.Claims.CodeHash = ""
	session.Claims.StateHash = ""

	delete(session.Claims.Extra, ClaimAuthorizedActor)
}

// TokenExchangeSubjectClientID returns the identifier of the client the subject token was issued to, and whether the
// subject token was a stored token. The result is false for subject tokens which were not stored, such as an ID Token.
func TokenExchangeSubjectClientID(session *Session) (clientID string, stored bool) {
	if session == nil {
		return "", false
	}

	requester := session.GetSubjectRequester()
	if requester == nil || requester.GetClient() == nil {
		return "", false
	}

	return requester.GetClient().GetID(), true
}

// TokenExchangeSubjectClientCredentials returns true if the stored subject token was issued via the client credentials
// grant.
func TokenExchangeSubjectClientCredentials(session *Session) (cc bool) {
	subject := tokenExchangeSubjectSession(session)

	return subject != nil && subject.ClientCredentials
}

func tokenExchangeSubjectSession(session *Session) *Session {
	if session == nil {
		return nil
	}

	requester := session.GetSubjectRequester()
	if requester == nil {
		return nil
	}

	subject, _ := requester.GetSession().(*Session)

	return subject
}
