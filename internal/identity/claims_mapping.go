// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"fmt"

	"github.com/authelia/authelia/v4/internal/expression"
)

// claimsMappedProvider applies the configured claims mapping to the identity another Provider established. The issuer
// and subject identify the account link and are never mapped.
type claimsMappedProvider struct {
	Provider

	mapping *expression.ClaimsMapping
	err     error
}

// newClaimsMappedProvider returns the provider unchanged unless a mapping is configured. Configuration validation
// rejects an invalid mapping, should one reach this point every sign in with the provider fails.
func newClaimsMappedProvider(provider Provider, mapping map[string]string) Provider {
	if len(mapping) == 0 {
		return provider
	}

	compiled, err := expression.NewClaimsMapping(mapping)

	return &claimsMappedProvider{Provider: provider, mapping: compiled, err: err}
}

// Complete implements Provider. A failure to evaluate the mapping fails the sign in rather than using the unmapped
// claims. The Authentication Method Reference expression receives the values of the policy of the provider.
func (p *claimsMappedProvider) Complete(ctx context.Context, request CompletionRequest) (claims *IdentityClaims, err error) {
	if p.err != nil {
		return nil, p.err
	}

	if claims, err = p.Provider.Complete(ctx, request); err != nil {
		return nil, err
	}

	raw, resolved := rawClaims(claims), p.Provider.AuthenticationMethodsReference(claims.AuthenticationMethodsReference)

	var disabled bool

	if disabled, err = p.mapping.Disabled(raw, resolved); err != nil {
		return nil, err
	}

	if disabled {
		return nil, fmt.Errorf("%w: issuer '%s' subject '%s'", ErrIdentityDisabled, claims.Issuer, claims.Subject)
	}

	values, amr, err := p.mapping.Evaluate(raw, resolved)
	if err != nil {
		return nil, err
	}

	if value, ok := values[expression.ClaimsMappingClaimPreferredUsername]; ok {
		claims.PreferredUsername = value
	}

	if value, ok := values[expression.ClaimsMappingClaimName]; ok {
		claims.Name = value
	}

	if value, ok := values[expression.ClaimsMappingClaimEmail]; ok {
		claims.Email = value
	}

	if p.mapping.MapsAMR() {
		claims.AuthenticationMethodsReference = amr
	}

	return claims, nil
}

// AuthenticationMethodsReference implements Provider. When the mapping includes the claim then Complete has already
// applied the policy of the provider, so the values are final.
func (p *claimsMappedProvider) AuthenticationMethodsReference(asserted []string) []string {
	if p.mapping != nil && p.mapping.MapsAMR() {
		return asserted
	}

	return p.Provider.AuthenticationMethodsReference(asserted)
}

// rawClaims returns the value of the claims variable of the expressions. Each source is always present, so that an
// expression can test for a claim without first testing for the source.
func rawClaims(claims *IdentityClaims) map[string]any {
	idToken, userInfo := claims.RawIDToken, claims.RawUserInfo

	if idToken == nil {
		idToken = map[string]any{}
	}

	if userInfo == nil {
		userInfo = map[string]any{}
	}

	sources := make(map[string]any, len(claims.RawProviders)+2)

	for name, source := range claims.RawProviders {
		sources[name] = source
	}

	sources[claimsMappingSourceIDToken] = idToken
	sources[claimsMappingSourceUserInfo] = userInfo

	return sources
}
