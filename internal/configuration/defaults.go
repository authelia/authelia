package configuration

var defaultsMapSource = map[string]any{
	"webauthn.selection_criteria.attachment":        "cross-platform",
	"webauthn.selection_criteria.discoverability":   "discouraged",
	"webauthn.selection_criteria.user_verification": "preferred",
}
