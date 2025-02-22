package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// MethodList is the list of available methods.
type MethodList = []string

// configurationBody the content returned by the configuration endpoint.
type configurationBody struct {
	AvailableMethods MethodList `json:"available_methods"`
}

// bodySignTOTPRequest is the  model of the request body of TOTP 2FA authentication endpoint.
type bodySignTOTPRequest struct {
	Token      string `json:"token" valid:"required"`
	TargetURL  string `json:"targetURL"`
	Workflow   string `json:"workflow"`
	WorkflowID string `json:"workflowID"`
}

type bodyRegisterTOTP struct {
	Algorithm string `json:"algorithm"`
	Length    int64  `json:"length"`
	Period    int    `json:"period"`
}

type bodyRegisterFinishTOTP struct {
	Token string `json:"token" valid:"required"`
}

// bodySignWebAuthnRequest is the  model of the request body of WebAuthn 2FA authentication endpoint.
type bodySignWebAuthnRequest struct {
	TargetURL  string `json:"targetURL"`
	Workflow   string `json:"workflow"`
	WorkflowID string `json:"workflowID"`

	Response json.RawMessage `json:"response"`
}

// bodyGETUserSessionElevate is the  model of the request body of the User Session Elevation PUT endpoint.
type bodyGETUserSessionElevate struct {
	RequireSecondFactor bool `json:"require_second_factor"`
	SkipSecondFactor    bool `json:"skip_second_factor"`
	CanSkipSecondFactor bool `json:"can_skip_second_factor"`
	Elevated            bool `json:"elevated"`
	Expires             int  `json:"expires"`
}

// bodyPOSTUserSessionElevate is the  model of the request body of the User Session Elevation PUT endpoint.
type bodyPOSTUserSessionElevate struct {
	DeleteID string `json:"delete_id"`
}

// bodyPUTUserSessionElevate is the  model of the request body of the User Session Elevation PUT endpoint.
type bodyPUTUserSessionElevate struct {
	OneTimeCode string `json:"otc"`
}

type bodyRegisterWebAuthnPUTRequest struct {
	Description string `json:"description"`
}

type bodyEditWebAuthnCredentialRequest struct {
	Description string `json:"description"`
}

// bodySignDuoRequest is the  model of the request body of Duo 2FA authentication endpoint.
type bodySignDuoRequest struct {
	TargetURL  string `json:"targetURL"`
	Passcode   string `json:"passcode"`
	Workflow   string `json:"workflow"`
	WorkflowID string `json:"workflowID"`
}

// bodyPreferred2FAMethod the selected 2FA method.
type bodyPreferred2FAMethod struct {
	Method string `json:"method" valid:"required"`
}

// bodyFirstFactorRequest represents the JSON body received by the endpoint.
type bodyFirstFactorRequest struct {
	Username       string `json:"username" valid:"required"`
	Password       string `json:"password" valid:"required"`
	Workflow       string `json:"workflow"`
	WorkflowID     string `json:"workflowID"`
	TargetURL      string `json:"targetURL"`
	RequestMethod  string `json:"requestMethod"`
	KeepMeLoggedIn *bool  `json:"keepMeLoggedIn"`
	// KeepMeLoggedIn: Cannot require this field because of https://github.com/asaskevich/govalidator/pull/329
	// TODO(c.michaud): add required validation once the above PR is merged.
}

// bodyFirstFactorRequest represents the JSON body received by the endpoint.
type bodyFirstFactorReauthenticateRequest struct {
	Password      string `json:"password" valid:"required"`
	Workflow      string `json:"workflow"`
	WorkflowID    string `json:"workflowID"`
	TargetURL     string `json:"targetURL"`
	RequestMethod string `json:"requestMethod"`
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

type bodyRequestPasswordResetDELETE struct {
	Token string `json:"token"`
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

type handlerAuthorizationConsent func(
	ctx *middlewares.AutheliaCtx, issuer *url.URL, client oidc.Client,
	userSession session.UserSession, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request,
	requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool)
