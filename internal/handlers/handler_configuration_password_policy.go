package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// PasswordPolicyConfigurationGET get the password policy configuration.
func PasswordPolicyConfigurationGET(ctx *middlewares.AutheliaCtx) {
	policyResponse := PasswordPolicyBody{
		Mode: "disabled",
	}

	if ctx.Configuration.PasswordPolicy.Standard.Enabled {
		policyResponse.Mode = "standard"
		policyResponse.MinLength = ctx.Configuration.PasswordPolicy.Standard.MinLength
		policyResponse.MaxLength = ctx.Configuration.PasswordPolicy.Standard.MaxLength
		policyResponse.RequireLowercase = ctx.Configuration.PasswordPolicy.Standard.RequireLowercase
		policyResponse.RequireUppercase = ctx.Configuration.PasswordPolicy.Standard.RequireUppercase
		policyResponse.RequireNumber = ctx.Configuration.PasswordPolicy.Standard.RequireNumber
		policyResponse.RequireSpecial = ctx.Configuration.PasswordPolicy.Standard.RequireSpecial
	} else if ctx.Configuration.PasswordPolicy.ZXCVBN.Enabled {
		policyResponse.Mode = "zxcvbn"
	}

	var err error

	if err = ctx.SetJSONBody(policyResponse); err != nil {
		ctx.Logger.Errorf("Unable to send password Policy: %s", err)
	}
}
