package configuration

var defaults = map[string]any{
	"webauthn.selection_criteria.attachment":        "cross-platform",
	"webauthn.selection_criteria.discoverability":   "discouraged",
	"webauthn.selection_criteria.user_verification": "preferred",
}

// Defaults returns a copy of the defaults.
func Defaults() map[string]any {
	values := map[string]any{}

	for k, v := range defaults {
		values[k] = v
	}

	return values
}
