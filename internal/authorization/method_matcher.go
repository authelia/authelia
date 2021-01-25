package authorization

import (
	"github.com/authelia/authelia/internal/utils"
)

func isMethodMatching(method string, methods []string) bool {
	// If there is no regexp patterns, it means that we match any path.
	if len(methods) == 0 {
		return true
	}

	if method == "" {
		return false
	}

	return utils.IsStringInSlice(method, methods)
}