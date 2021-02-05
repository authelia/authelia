package authorization

import (
	"github.com/authelia/authelia/internal/utils"
)

// isMethodMatching checks if the request method matches a method in the rule.
func isMethodMatching(method string, methods []string) bool {
	// If there are no defined methods, it means that we match all methods.
	if len(methods) == 0 {
		return true
	}

	if method == "" {
		return false
	}

	return utils.IsStringInSlice(method, methods)
}
