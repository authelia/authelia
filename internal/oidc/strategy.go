// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	oauthelia2 "authelia.com/provider/oauth2"
)

// Audience and resource strategy names used in the configuration.
const (
	StrategyExact    = "exact"
	StrategyPrefix   = "prefix"
	StrategyWildcard = "wildcard"
)

// AudienceStrategyNames is the list of valid audience strategy names.
var AudienceStrategyNames = []string{StrategyExact, StrategyWildcard}

// ResourceStrategyNames is the list of valid resource strategy names.
var ResourceStrategyNames = []string{StrategyExact, StrategyPrefix, StrategyWildcard}

// NewAudienceStrategy returns the oauthelia2.AudienceStrategy for a configured strategy name. It returns nil for an
// empty or unknown name so the caller falls back to the next strategy.
func NewAudienceStrategy(name string) oauthelia2.AudienceStrategy {
	switch name {
	case StrategyExact:
		return oauthelia2.ExactAudienceStrategy
	case StrategyWildcard:
		return oauthelia2.WildcardAudienceStrategy
	default:
		return nil
	}
}

// NewResourceStrategy returns the oauthelia2.ResourceStrategy for a configured strategy name. It returns nil for an
// empty or unknown name so the caller falls back to the next strategy.
func NewResourceStrategy(name string) oauthelia2.ResourceStrategy {
	switch name {
	case StrategyExact:
		return oauthelia2.ExactResourceStrategy
	case StrategyPrefix:
		return oauthelia2.PrefixResourceStrategy
	case StrategyWildcard:
		return oauthelia2.WildcardResourceStrategy
	default:
		return nil
	}
}
