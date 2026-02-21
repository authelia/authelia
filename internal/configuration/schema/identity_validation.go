package schema

import (
	"time"
)

// IdentityValidation represents the configuration for identity verification actions/flows.
type IdentityValidation struct {
	ResetPassword   IdentityValidationResetPassword   `koanf:"reset_password" yaml:"reset_password,omitempty" toml:"reset_password,omitempty" json:"reset_password,omitempty" jsonschema:"title=Reset Password" jsonschema_description:"Identity Validation options for the Reset Password flow."`
	ElevatedSession IdentityValidationElevatedSession `koanf:"elevated_session" yaml:"elevated_session,omitempty" toml:"elevated_session,omitempty" json:"elevated_session,omitempty" jsonschema:"title=Elevated Session" jsonschema_description:"Identity Validation options for obtaining an Elevated Session for flows such as the Credential Management flows."`
}

// IdentityValidationResetPassword represents the tunable aspects of the reset password identity verification action/flow.
type IdentityValidationResetPassword struct {
	JWTExpiration time.Duration `koanf:"jwt_lifespan" yaml:"jwt_lifespan,omitempty" toml:"jwt_lifespan,omitempty" json:"jwt_lifespan,omitempty" jsonschema:"title=JWT Lifespan,default=5 minutes" jsonschema_description:"The lifespan of the JSON Web Token after it's initially generated after which it's considered invalid."`
	JWTAlgorithm  string        `koanf:"jwt_algorithm" yaml:"jwt_algorithm,omitempty" toml:"jwt_algorithm,omitempty" json:"jwt_algorithm,omitempty" jsonschema:"title=JWT Algorithm,default=HS256,enum=HS256,enum=HS384,enum=HS512" jsonschema_description:"The JSON Web Token Algorithm (JWA) used to sign the Reset Password flow JSON Web Token's."`
	JWTSecret     string        `koanf:"jwt_secret" yaml:"jwt_secret,omitempty" toml:"jwt_secret,omitempty" json:"jwt_secret,omitempty" jsonschema:"title=JWT Secret" jsonschema_description:"The secret key used to sign the Reset Password flow JSON Web Token's."` //nolint:gosec
}

// IdentityValidationElevatedSession represents the tunable aspects of the credential control identity verification action/flow.
type IdentityValidationElevatedSession struct {
	CodeLifespan        time.Duration `koanf:"code_lifespan" yaml:"code_lifespan,omitempty" toml:"code_lifespan,omitempty" json:"code_lifespan,omitempty" jsonschema:"title=Code Lifespan,default=5 minutes" jsonschema_description:"The lifespan of the randomly generated One Time Code after which it's considered invalid."`
	ElevationLifespan   time.Duration `koanf:"elevation_lifespan" yaml:"elevation_lifespan,omitempty" toml:"elevation_lifespan,omitempty" json:"elevation_lifespan,omitempty" jsonschema:"title=Elevation Lifespan,default=10 minutes" jsonschema_description:"The lifespan of the elevation after initially validating the One-Time Code before it expires."`
	Characters          int           `koanf:"characters" yaml:"characters" toml:"characters" json:"characters" jsonschema:"title=OTP Characters,minimum=6,maximum=12,default=8" jsonschema_description:"Number of characters in the generated OTP codes."`
	RequireSecondFactor bool          `koanf:"require_second_factor" yaml:"require_second_factor" toml:"require_second_factor" json:"require_second_factor" jsonschema:"title=Require Second Factor,default=false" jsonschema_description:"Requires the user use a second factor if they have any known second factor methods."`
	SkipSecondFactor    bool          `koanf:"skip_second_factor" yaml:"skip_second_factor" toml:"skip_second_factor" json:"skip_second_factor" jsonschema:"title=Skip Second Factor,default=false" jsonschema_description:"Skips the primary identity verification process if the user has authenticated with a second factor."`
}

// DefaultIdentityValidation has the default values for the IdentityValidation configuration.
var DefaultIdentityValidation = IdentityValidation{
	ResetPassword: IdentityValidationResetPassword{
		JWTExpiration: time.Minute * 5,
		JWTAlgorithm:  "HS256",
	},
	ElevatedSession: IdentityValidationElevatedSession{
		CodeLifespan:      time.Minute * 5,
		ElevationLifespan: time.Minute * 10,
		Characters:        8,
	},
}
