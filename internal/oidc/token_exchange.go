// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"github.com/google/uuid"
)

// HydrateTokenExchangeSessionWithRequestingClient re-establishes the identity of the client performing an RFC 8693
// Token Exchange on the session used to issue the exchanged token.
//
// The RFC 8693 token type handlers resolve the 'subject_token' (and the 'actor_token') back to the request it was
// issued for by calling the storage layer with the LIVE exchange request's session. model.OAuth2Session.ToRequest
// unmarshals the stored session JSON into that same pointer, so every persisted field of the exchange session is
// replaced by the subject token's, including the identity of the client the subject token was issued to and the whole
// embedded ID Token claims block.
//
// Without this the exchanged token would be attributed to the wrong client: an ID Token minted from an access or
// refresh token would carry the original client as its 'azp', and would inherit that client's audience, nonce,
// expiry and the 'at_hash', 'c_hash' and 's_hash' bindings of an authorization response it was never part of.
//
// Only the fields identifying the party the token is issued to, and the bindings which cannot survive the exchange,
// are reset. The subject's identity and the claims describing their authentication are deliberately retained as they
// are precisely what the exchange carries across.
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
}
