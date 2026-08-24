package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// MethodList is the list of available methods.
type MethodList = []string

type configurationBody struct {
	AvailableMethods       MethodList `json:"available_methods"`
	PasswordChangeDisabled bool       `json:"password_change_disabled"`
	PasswordResetDisabled  bool       `json:"password_reset_disabled"`
}

type bodySignTOTPRequest struct {
	Token     string `json:"token" valid:"required"`
	TargetURL string `json:"targetURL"`
	FlowID    string `json:"flowID"`
	Flow      string `json:"flow"`
	SubFlow   string `json:"subflow"`
	UserCode  string `json:"userCode"`
}

type bodyRegisterTOTP struct {
	Algorithm string `json:"algorithm"`
	Length    int64  `json:"length"`
	Period    int    `json:"period"`
}

type bodyRegisterFinishTOTP struct {
	Token string `json:"token" valid:"required"`
}

type bodySignWebAuthnRequest struct {
	TargetURL string `json:"targetURL"`
	FlowID    string `json:"flowID"`
	Flow      string `json:"flow"`
	SubFlow   string `json:"subflow"`
	UserCode  string `json:"userCode"`

	Response json.RawMessage `json:"response"`
}

type bodySignPasskeyRequest struct {
	TargetURL      string `json:"targetURL"`
	RequestMethod  string `json:"requestMethod"`
	KeepMeLoggedIn *bool  `json:"keepMeLoggedIn"`
	FlowID         string `json:"flowID"`
	Flow           string `json:"flow"`
	SubFlow        string `json:"subflow"`
	UserCode       string `json:"userCode"`

	Response json.RawMessage `json:"response"`
}

type bodyGETUserSessionElevate struct {
	RequireSecondFactor bool `json:"require_second_factor"`
	SkipSecondFactor    bool `json:"skip_second_factor"`
	CanSkipSecondFactor bool `json:"can_skip_second_factor"`
	FactorKnowledge     bool `json:"factor_knowledge"`
	Elevated            bool `json:"elevated"`
	Expires             int  `json:"expires"`
}

type bodyPOSTUserSessionElevate struct {
	DeleteID string `json:"delete_id"`
}

type bodyPUTUserSessionElevate struct {
	OneTimeCode string `json:"otc"`
}

type bodyRegisterWebAuthnPUTRequest struct {
	Description string `json:"description"`
}

type bodyEditWebAuthnCredentialRequest struct {
	Description string `json:"description"`
}

type bodySignDuoRequest struct {
	TargetURL string `json:"targetURL"`
	Passcode  string `json:"passcode"`
	FlowID    string `json:"flowID"`
	Flow      string `json:"flow"`
	SubFlow   string `json:"subflow"`
	UserCode  string `json:"userCode"`
}

type bodyPreferred2FAMethod struct {
	Method string `json:"method" valid:"required"`
}

type bodyFirstFactorRequest struct {
	Username       string `json:"username" valid:"required"`
	Password       string `json:"password" valid:"required"`
	TargetURL      string `json:"targetURL"`
	RequestMethod  string `json:"requestMethod"`
	KeepMeLoggedIn *bool  `json:"keepMeLoggedIn"`
	FlowID         string `json:"flowID"`
	Flow           string `json:"flow"`
	SubFlow        string `json:"subflow"`
	UserCode       string `json:"userCode"`
}

type bodySecondFactorPasswordRequest struct {
	Password  string `json:"password" valid:"required"`
	TargetURL string `json:"targetURL"`
	FlowID    string `json:"flowID"`
	Flow      string `json:"flow"`
	SubFlow   string `json:"subflow"`
	UserCode  string `json:"userCode"`
}

type bodyFirstFactorReauthenticateRequest struct {
	Password      string `json:"password" valid:"required"`
	TargetURL     string `json:"targetURL"`
	RequestMethod string `json:"requestMethod"`
	FlowID        string `json:"flowID"`
	Flow          string `json:"flow"`
	SubFlow       string `json:"subflow"`
	UserCode      string `json:"userCode"`
}

type checkURIWithinDomainRequestBody struct {
	URI string `json:"uri"`
}

type checkURIWithinDomainResponseBody struct {
	OK bool `json:"ok"`
}

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

// DuoDevicesResponse represents all available user devices and methods, the user's preferred device and method selections, and an optional enrollment url.
type DuoDevicesResponse struct {
	Result          string      `json:"result" valid:"required"`
	Devices         []DuoDevice `json:"devices,omitempty"`
	EnrollURL       string      `json:"enroll_url,omitempty"`
	PreferredDevice string      `json:"preferred_device,omitempty"`
	PreferredMethod string      `json:"preferred_method,omitempty"`
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
	FactorKnowledge       bool                 `json:"factor_knowledge"`
	DefaultRedirectionURL string               `json:"default_redirection_url,omitempty"`
}

type resetPasswordStep1RequestBody struct {
	Username string `json:"username"`
}

type resetPasswordStep2RequestBody struct {
	Password string `json:"password"`
}

type bodyRequestPasswordResetDELETE struct {
	Token string `json:"token"`
}

type changePasswordRequestBody struct {
	Username    string `json:"username"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
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
	userSession session.UserSession, details *authentication.UserDetailsExtended, subject uuid.UUID,
	rw http.ResponseWriter, r *http.Request,
	requester oauthelia2.Requester) (consent *model.OAuth2ConsentSession, handled bool)
