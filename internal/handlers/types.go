package handlers

import (
	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
)

// MethodList is the list of available methods.
type MethodList = []string

// configurationBody the content returned by the configuration endpoint.
type configurationBody struct {
	AvailableMethods MethodList `json:"available_methods"`
}

// signTOTPRequestBody model of the request body received by TOTP authentication endpoint.
type signTOTPRequestBody struct {
	Token     string `json:"token" valid:"required"`
	TargetURL string `json:"targetURL"`
	Workflow  string `json:"workflow"`
}

// signWebauthnRequestBody model of the request body of Webauthn authentication endpoint.
type signWebauthnRequestBody struct {
	TargetURL string `json:"targetURL"`
	Workflow  string `json:"workflow"`
}

type signDuoRequestBody struct {
	TargetURL string `json:"targetURL"`
	Passcode  string `json:"passcode"`
	Workflow  string `json:"workflow"`
}

// preferred2FAMethodBody the selected 2FA method.
type preferred2FAMethodBody struct {
	Method string `json:"method" valid:"required"`
}

// firstFactorRequestBody represents the JSON body received by the endpoint.
type firstFactorRequestBody struct {
	Username       string `json:"username" valid:"required"`
	Password       string `json:"password" valid:"required"`
	TargetURL      string `json:"targetURL"`
	Workflow       string `json:"workflow"`
	RequestMethod  string `json:"requestMethod"`
	KeepMeLoggedIn *bool  `json:"keepMeLoggedIn"`
	// KeepMeLoggedIn: Cannot require this field because of https://github.com/asaskevich/govalidator/pull/329
	// TODO(c.michaud): add required validation once the above PR is merged.
}

// checkURIWithinDomainRequestBody represents the JSON body received by the endpoint checking if an URI is within
// the configured domain.
type checkURIWithinDomainRequestBody struct {
	URI string `json:"uri"`
}

type checkURIWithinDomainResponseBody struct {
	OK bool `json:"ok"`
}

// redirectResponse represent the response sent by the first factor endpoint
// when a redirection URL has been provided.
type redirectResponse struct {
	Redirect string `json:"redirect"`
}

// TOTPKeyResponse is the model of response that is sent to the client up successful identity verification.
type TOTPKeyResponse struct {
	Base32Secret string `json:"base32_secret"`
	OTPAuthURL   string `json:"otpauth_url"`
}

// DuoDeviceBody the selected Duo device and method.
type DuoDeviceBody struct {
	Device string `json:"device" valid:"required"`
	Method string `json:"method" valid:"required"`
}

// DuoDevice represents Duo devices and methods.
type DuoDevice struct {
	Device       string   `json:"device"`
	DisplayName  string   `json:"display_name"`
	Capabilities []string `json:"capabilities"`
}

// DuoDevicesResponse represents all available user devices and methods as well as an optional enrollment url.
type DuoDevicesResponse struct {
	Result    string      `json:"result" valid:"required"`
	Devices   []DuoDevice `json:"devices,omitempty"`
	EnrollURL string      `json:"enroll_url,omitempty"`
}

// DuoSignResponse represents a result of the preauth and or auth call with further optional info.
type DuoSignResponse struct {
	Result    string      `json:"result" valid:"required"`
	Devices   []DuoDevice `json:"devices,omitempty"`
	Redirect  string      `json:"redirect,omitempty"`
	EnrollURL string      `json:"enroll_url,omitempty"`
}

// StateResponse represents the response sent by the state endpoint.
type StateResponse struct {
	Username              string               `json:"username"`
	AuthenticationLevel   authentication.Level `json:"authentication_level"`
	DefaultRedirectionURL string               `json:"default_redirection_url"`
}

// resetPasswordStep1RequestBody model of the reset password (step1) request body.
type resetPasswordStep1RequestBody struct {
	Username string `json:"username"`
}

// resetPasswordStep2RequestBody model of the reset password (step2) request body.
type resetPasswordStep2RequestBody struct {
	Password string `json:"password"`
}

// PasswordPolicyBody represents the response sent by the password reset step 2.
type PasswordPolicyBody struct {
	Mode             string `json:"mode"`
	MinLength        int    `json:"min_length"`
	MaxLength        int    `json:"max_length"`
	MinScore         int    `json:"min_score"`
	RequireUppercase bool   `json:"require_uppercase"`
	RequireLowercase bool   `json:"require_lowercase"`
	RequireNumber    bool   `json:"require_number"`
	RequireSpecial   bool   `json:"require_special"`
}

// AuthnType is an auth type.
type AuthnType int

const (
	// AuthnTypeNone is a nil Authentication AuthnType.
	AuthnTypeNone AuthnType = iota

	// AuthnTypeCookie is an Authentication AuthnType based on the Cookie header.
	AuthnTypeCookie

	// AuthnTypeProxyAuthorization is an Authentication AuthnType based on the Proxy-Authorization header.
	AuthnTypeProxyAuthorization

	// AuthnTypeAuthorization is an Authentication AuthnType based on the Authorization header.
	AuthnTypeAuthorization
)

// Authn is authentication.
type Authn struct {
	Username string
	Method   string

	Details authentication.UserDetails
	Level   authentication.Level
	Object  authorization.Object
	Type    AuthnType
}
