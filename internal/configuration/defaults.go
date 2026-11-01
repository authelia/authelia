// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

var defaults = newDefaults()

func newDefaults() (values map[string]any) {
	values = map[string]any{
		"regulation.max_retries":                        3,
		"webauthn.selection_criteria.discoverability":   "preferred",
		"webauthn.selection_criteria.user_verification": "preferred",
		"webauthn.metadata.cache_policy":                "strict",
	}

	limits := schema.ServerEndpointRateLimits{}

	for _, endpoint := range limits.Endpoints() {
		values["server.endpoints.rate_limits."+endpoint.Name+".enable"] = true
	}

	return values
}

// Defaults returns a copy of the defaults.
func Defaults() map[string]any {
	values := map[string]any{}

	for k, v := range defaults {
		values[k] = v
	}

	return values
}
