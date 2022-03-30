package validator

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateDefault validates the Default configuration.
func ValidateDefault(config schema.DefaultConfiguration, validator *schema.StructValidator) {
	if config.UserSecondFactorMethod != "" && !utils.IsStringInSlice(config.UserSecondFactorMethod, validDefaultUserSecondFactorMethods) {
		validator.Push(fmt.Errorf(errFmtDefaultInvalidMethod, config.UserSecondFactorMethod, strings.Join(validDefaultUserSecondFactorMethods, "', '")))
	}
}
