package configuration

var defaults = map[string]any{
	"webauthn.selection_criteria.attachment":              "cross-platform",
	"webauthn.selection_criteria.discoverability":         "preferred",
	"webauthn.selection_criteria.user_verification":       "preferred",
	"webauthn.metadata.validate_trust_anchor":             true,
	"webauthn.metadata.validate_entry":                    true,
	"webauthn.metadata.validate_entry_permit_zero_aaguid": false,
	"webauthn.metadata.validate_status":                   true,
}

// Defaults returns a copy of the defaults.
func Defaults() map[string]any {
	values := map[string]any{}

	for k, v := range defaults {
		values[k] = v
	}

	return values
}
