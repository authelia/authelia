package authorization

import (
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

// isMethodMatching check whether request method is in the rules methods
func isMethodMatching(method []byte, methods []string) (matching bool) {
	log := logging.Logger()
	// If no method is provided in the rule, we match any method
	if len(methods) == 0 {
		log.Trace("Method matching skipped as no methods have been defined for the access_control rule")
		return true
	}

	if method == nil {
		log.Warn("Rules have been configured with methods but your proxy does not seem to be " +
			"sending the X-Forwarded-Method header, these rules will effectively be disabled")
		return false
	}

	methodStr := string(method)
	if !utils.IsStringInSlice(methodStr, schema.HTTPRequestMethods) {
		log.Warnf("An invalid X-Forwarded-Method header was received, its value was %s", methodStr)
		return false
	}
	return utils.IsStringInSlice(methodStr, methods)
}
