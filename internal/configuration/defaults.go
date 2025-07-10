package configuration

var defaults = map[string]any{
	"webauthn.selection_criteria.discoverability":                  "preferred",
	"webauthn.selection_criteria.user_verification":                "preferred",
	"webauthn.metadata.cache_policy":                               "strict",
	"server.endpoints.rate_limits.reset_password_start.enable":     true,
	"server.endpoints.rate_limits.reset_password_finish.enable":    true,
	"server.endpoints.rate_limits.second_factor_totp.enable":       true,
	"server.endpoints.rate_limits.second_factor_duo.enable":        true,
	"server.endpoints.rate_limits.session_elevation_start.enable":  true,
	"server.endpoints.rate_limits.session_elevation_finish.enable": true,
}

// Defaults returns a copy of the defaults.
func Defaults() map[string]any {
	values := map[string]any{}

	for k, v := range defaults {
		values[k] = v
	}

	return values
}
