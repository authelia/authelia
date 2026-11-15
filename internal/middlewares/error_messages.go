// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

// ErrorMessage is a user facing API error with a stable machine readable code and a human readable message.
type ErrorMessage struct {
	Code    string
	Message string
}

// Error messages returned in KO responses.
var (
	ErrorMessageOperationFailed                 = ErrorMessage{Code: "operation_failed", Message: "Operation failed."}
	ErrorMessageAuthenticationFailed            = ErrorMessage{Code: "authentication_failed", Message: "Authentication failed. Check your credentials."}
	ErrorMessageMFAValidationFailed             = ErrorMessage{Code: "mfa_validation_failed", Message: "Authentication failed, please retry later."}
	ErrorMessageRateLimited                     = ErrorMessage{Code: "rate_limited", Message: "Too Many Requests"}
	ErrorMessagePasswordPolicy                  = ErrorMessage{Code: "password_policy", Message: "Your supplied password does not meet the password policy requirements."}
	ErrorMessagePasswordBackendComplexity       = ErrorMessage{Code: "password_backend_complexity", Message: "0000052D."}
	ErrorMessagePasswordIncorrect               = ErrorMessage{Code: "password_incorrect", Message: "Incorrect Password"}
	ErrorMessagePasswordResetFailed             = ErrorMessage{Code: "password_reset_failed", Message: "Unable to reset your password."}
	ErrorMessagePasswordChangeFailed            = ErrorMessage{Code: "password_change_failed", Message: "Unable to change your password."}
	ErrorMessageTOTPOptionsFailed               = ErrorMessage{Code: "totp_options_failed", Message: "Unable to retrieve TOTP registration options."}
	ErrorMessageTOTPRegisterFailed              = ErrorMessage{Code: "totp_register_failed", Message: "Unable to set up one-time password."}
	ErrorMessageTOTPRegisterSessionDeleteFailed = ErrorMessage{Code: "totp_register_session_delete_failed", Message: "Unable to delete one-time password registration session."}
	ErrorMessageTOTPDeleteFailed                = ErrorMessage{Code: "totp_delete_failed", Message: "Unable to delete one-time password."}
	ErrorMessageTOTPConfigurationNotFound       = ErrorMessage{Code: "totp_configuration_not_found", Message: "Could not find TOTP Configuration for user."}
	ErrorMessageWebAuthnRegisterFailed          = ErrorMessage{Code: "webauthn_register_failed", Message: "Unable to register your security key."}
	ErrorMessageWebAuthnDuplicateName           = ErrorMessage{Code: "webauthn_duplicate_name", Message: "Another one of your security keys is already registered with that display name."}
	ErrorMessageIdentityTokenExpired            = ErrorMessage{Code: "identity_token_expired", Message: "The identity verification token has expired"}
	ErrorMessageIdentityTokenNotYetValid        = ErrorMessage{Code: "identity_token_not_yet_valid", Message: "The identity verification token is only valid in the future"}
	ErrorMessageIdentityTokenInvalidSignature   = ErrorMessage{Code: "identity_token_invalid_signature", Message: "The identity verification token has an invalid signature"}
	ErrorMessageIdentityTokenUsed               = ErrorMessage{Code: "identity_token_used", Message: "The identity verification token has already been used"}
)

// ErrorMessages is every ErrorMessage returned by the API.
var ErrorMessages = []ErrorMessage{
	ErrorMessageOperationFailed,
	ErrorMessageAuthenticationFailed,
	ErrorMessageMFAValidationFailed,
	ErrorMessageRateLimited,
	ErrorMessagePasswordPolicy,
	ErrorMessagePasswordBackendComplexity,
	ErrorMessagePasswordIncorrect,
	ErrorMessagePasswordResetFailed,
	ErrorMessagePasswordChangeFailed,
	ErrorMessageTOTPOptionsFailed,
	ErrorMessageTOTPRegisterFailed,
	ErrorMessageTOTPRegisterSessionDeleteFailed,
	ErrorMessageTOTPDeleteFailed,
	ErrorMessageTOTPConfigurationNotFound,
	ErrorMessageWebAuthnRegisterFailed,
	ErrorMessageWebAuthnDuplicateName,
	ErrorMessageIdentityTokenExpired,
	ErrorMessageIdentityTokenNotYetValid,
	ErrorMessageIdentityTokenInvalidSignature,
	ErrorMessageIdentityTokenUsed,
}
