// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

export type ErrorCode =
    | "authentication_failed"
    | "group_already_exists"
    | "group_create_failed"
    | "group_name_required"
    | "group_not_found"
    | "identity_token_expired"
    | "identity_token_invalid_signature"
    | "identity_token_not_yet_valid"
    | "identity_token_used"
    | "invalid_request"
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
    | "user_already_exists"
    | "user_create_incomplete"
    | "user_delete_failed"
    | "user_email_invalid"
    | "user_family_name_required"
    | "user_modify_failed"
    | "user_not_found"
    | "user_password_modify_unsupported"
    | "user_password_reset_email_failed"
    | "user_update_mask_invalid"
    | "user_username_required"
    | "webauthn_duplicate_name"
    | "webauthn_register_failed";

export const ErrorCodeTranslationKeys: Record<ErrorCode, string> = {
    authentication_failed: "Authentication failed, check your credentials",
    group_already_exists: "Group already exists",
    group_create_failed: "Unable to create the group",
    group_name_required: "Group name is required",
    group_not_found: "Group not found",
    identity_token_expired: "The identity verification token has expired",
    identity_token_invalid_signature: "The identity verification token has an invalid signature",
    identity_token_not_yet_valid: "The identity verification token is only valid in the future",
    identity_token_used: "The identity verification token has already been used",
    invalid_request: "The request is invalid",
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
    user_already_exists: "User already exists",
    user_create_incomplete: "The user has been created but metadata creation failed",
    user_delete_failed: "Unable to delete the user",
    user_email_invalid: "Invalid email address",
    user_family_name_required: "Last name is required",
    user_modify_failed: "Unable to modify the user",
    user_not_found: "User not found",
    user_password_modify_unsupported: "Password modification is not supported by this endpoint",
    user_password_reset_email_failed: "Unable to send the password reset email",
    user_update_mask_invalid: "The update mask is missing or contains an unsupported field",
    user_username_required: "Username is required",
    webauthn_duplicate_name: "Another one of your security keys is already registered with that display name",
    webauthn_register_failed: "Unable to register your security key",
};

export function isErrorCode(code: unknown): code is ErrorCode {
    return typeof code === "string" && Object.keys(ErrorCodeTranslationKeys).includes(code);
}

export interface Translate {
    (key: string): string;
    (key: string, options: Record<string, unknown>): string;
}

export function translateErrorCode(translate: Translate, code: string | undefined, fallback: string): string {
    if (!isErrorCode(code)) {
        return fallback;
    }

    return translate(ErrorCodeTranslationKeys[code], { ns: "portal" });
}
