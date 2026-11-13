// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"fmt"

	"authelia.com/provider/jose"
	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/rfc8693"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewRFC8693TokenTypes builds the token types of the configured RFC8693 token exchange, keyed by their URN. The static
// types, including the Identity Assertion JWT Authorization Grant type which enables issuing grants, are built for each
// configured URN and the Transaction Token type from the Transaction Tokens configuration. The access, refresh, and ID
// token types are used when none are configured.
func NewRFC8693TokenTypes(config *schema.IdentityProvidersOpenIDConnectTokenExchange, issuer *Issuer) (types map[string]oauthelia2.RFC8693TokenType) {
	names := config.TokenTypes

	if len(names) == 0 {
		names = []string{TokenTypeAccessToken, TokenTypeRefreshToken, TokenTypeIDToken}
	}

	types = make(map[string]oauthelia2.RFC8693TokenType, len(names))

	for _, name := range names {
		if name == TokenTypeTxnToken {
			types[name] = newTxnTokenType(&config.TransactionTokens, issuer)

			continue
		}

		types[name] = &rfc8693.DefaultTokenType{Name: name}
	}

	return types
}

// NewIssuerKeyFunc returns a jwt.Keyfunc which resolves the public key of the issuer that signed a token.
func NewIssuerKeyFunc(issuer *Issuer) (keyfunc jwt.Keyfunc) {
	return func(token *jwt.Token) (key any, err error) {
		if issuer == nil {
			return nil, fmt.Errorf("the issuer is not available")
		}

		var jwk *jose.JSONWebKey

		if jwk, err = issuer.GetIssuerStrictJWK(context.Background(), token.KeyID, string(token.SignatureAlgorithm), KeyUseSignature); err != nil {
			return nil, err
		}

		return jwk.Public().Key, nil
	}
}

func defaultRequestedTokenType(config *schema.IdentityProvidersOpenIDConnectTokenExchange) string {
	if config.DefaultRequestedTokenType == "" {
		return TokenTypeAccessToken
	}

	return config.DefaultRequestedTokenType
}

func newTxnTokenType(config *schema.IdentityProvidersOpenIDConnectTransactionTokens, issuer *Issuer) (tokenType *rfc8693.TxnTokenType) {
	domain := &rfc8693.TxnTokenTrustDomain{
		Issuer:     config.Issuer,
		Expiry:     config.Lifespan,
		SigningAlg: config.SigningAlgorithm,
	}

	if config.CopyRequestParameters {
		domain.TranscribeClaims = rfc8693.CopyTxnTokenRequestParameters
	}

	if config.Replacement.Enabled {
		domain.Replacement = &rfc8693.TxnTokenReplacement{
			KeyFunc:        NewIssuerKeyFunc(issuer),
			MaxChainLength: config.Replacement.MaxChainLength,
		}
	}

	return &rfc8693.TxnTokenType{
		Name:    TokenTypeTxnToken,
		Domains: map[string]*rfc8693.TxnTokenTrustDomain{config.TrustDomain: domain},
	}
}
