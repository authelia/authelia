// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

export type ErrorCode =
    | "authentication_failed"
    | "identity_token_expired"
    | "identity_token_invalid_signature"
    | "identity_token_not_yet_valid"
    | "identity_token_used"
    | "mfa_validation_failed"
    | "operation_failed"
    | "password_backend_complexity"
    | "password_change_failed"
    | "password_incorrect"
    | "password_policy"
    | "password_reset_failed"
    | "rate_limited"
    | "totp_configuration_not_found"
    | "totp_delete_failed"
    | "totp_options_failed"
    | "totp_register_failed"
    | "totp_register_session_delete_failed"
    | "webauthn_duplicate_name"
    | "webauthn_register_failed";

export const ErrorCodeTranslationKeys: Record<ErrorCode, string> = {
    authentication_failed: "Authentication failed, check your credentials",
    identity_token_expired: "The identity verification token has expired",
    identity_token_invalid_signature: "The identity verification token has an invalid signature",
    identity_token_not_yet_valid: "The identity verification token is only valid in the future",
    identity_token_used: "The identity verification token has already been used",
    mfa_validation_failed: "Authentication failed, please retry later",
    operation_failed: "Operation failed",
    password_backend_complexity: "Your supplied password does not meet the password policy requirements",
    password_change_failed: "Unable to change your password",
    password_incorrect: "Incorrect password",
    password_policy: "Your supplied password does not meet the password policy requirements",
    password_reset_failed: "Unable to reset your password",
    rate_limited: "You have made too many requests",
    totp_configuration_not_found: "Could not find the One-Time Password configuration",
    totp_delete_failed: "Unable to delete the One-Time Password",
    totp_options_failed: "Unable to retrieve the One-Time Password registration options",
    totp_register_failed: "Unable to set up the One-Time Password",
    totp_register_session_delete_failed: "Unable to delete the One-Time Password registration session",
    webauthn_duplicate_name: "Another one of your security keys is already registered with that display name",
    webauthn_register_failed: "Unable to register your security key",
};

export function isErrorCode(code: unknown): code is ErrorCode {
    return typeof code === "string" && Object.keys(ErrorCodeTranslationKeys).includes(code);
}

export function translateErrorCode(
    translate: (_key: string, _options?: any) => string,
    code: string | undefined,
    fallback: string,
): string {
    if (!isErrorCode(code)) {
        return fallback;
    }

    return translate(ErrorCodeTranslationKeys[code], { ns: "portal" });
}
