// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { AxiosResponse } from "axios";

import { getBasePath } from "@utils/BasePath";

const basePath = getBasePath();

export const OpenIDConnectConsentPath = basePath + "/api/oidc/consent";
export const OpenIDConnectDeviceAuthorizationPath = basePath + "/api/oidc/device-authorization";

export const FirstFactorPath = basePath + "/api/firstfactor";
export const FirstFactorPasskeyPath = basePath + "/api/firstfactor/passkey";
export const FirstFactorPasskeyConditionalPath = basePath + "/api/firstfactor/passkey/conditional";
export const FirstFactorReauthenticatePath = basePath + "/api/firstfactor/reauthenticate";

export const TOTPRegistrationPath = basePath + "/api/secondfactor/totp/register";
export const TOTPConfigurationPath = basePath + "/api/secondfactor/totp";

export const WebAuthnRegistrationPath = basePath + "/api/secondfactor/webauthn/credential/register";
export const WebAuthnAssertionPath = basePath + "/api/secondfactor/webauthn";
export const WebAuthnCredentialsPath = basePath + "/api/secondfactor/webauthn/credentials";
export const WebAuthnCredentialPath = basePath + "/api/secondfactor/webauthn/credential";

export const InitiateDuoDeviceSelectionPath = basePath + "/api/secondfactor/duo_devices";
export const CompleteDuoDeviceSelectionPath = basePath + "/api/secondfactor/duo_device";

export const CompletePushNotificationSignInPath = basePath + "/api/secondfactor/duo";
export const CompleteTOTPSignInPath = basePath + "/api/secondfactor/totp";
export const CompletePasswordSignInPath = basePath + "/api/secondfactor/password";

export const InitiateResetPasswordPath = basePath + "/api/reset-password/identity/start";
export const CompleteResetPasswordPath = basePath + "/api/reset-password/identity/finish";

export const ChangePasswordPath = basePath + "/api/change-password";

export const ResetPasswordPath = basePath + "/api/reset-password";
export const ChecksSafeRedirectionPath = basePath + "/api/checks/safe-redirection";

export const LogoutPath = basePath + "/api/logout";
export const StatePath = basePath + "/api/state";
export const UserInfoPath = basePath + "/api/user/info";
export const UserInfo2FAMethodPath = basePath + "/api/user/info/2fa_method";
export const UserSessionElevationPath = basePath + "/api/user/session/elevation";

export const ConfigurationPath = basePath + "/api/configuration";
export const PasswordPolicyConfigurationPath = basePath + "/api/configuration/password-policy";

export const LocaleInformationPath = basePath + "/locales";

export interface AuthenticationErrorResponse extends ErrorResponse {
    authentication: boolean;
    elevation: boolean;
}

export interface ErrorResponse {
    status: "KO";
    message: string;
}

export interface Response<T> extends OKResponse {
    data: T;
}

export interface OptionalDataResponse<T> extends OKResponse {
    data?: T;
}

export interface OKResponse {
    status: "OK";
}

export type AuthenticationResponse<T> = AuthenticationErrorResponse | Response<T>;
export type AuthenticationOKResponse = AuthenticationErrorResponse | OKResponse;
export type OptionalDataServiceResponse<T> = ErrorResponse | OptionalDataResponse<T>;
export type ServiceResponse<T> = ErrorResponse | Response<T>;

function toErrorResponse<T>(resp: AxiosResponse<ServiceResponse<T>>): ErrorResponse | undefined {
    if (resp.data && "status" in resp.data && resp.data["status"] === "KO") {
        return resp.data;
    }

    return undefined;
}

export function toData<T>(resp: AxiosResponse<ServiceResponse<T>>): T | undefined {
    if (resp.data && "status" in resp.data && resp.data["status"] === "OK") {
        return resp.data.data;
    }

    return undefined;
}

function hasError(err: ErrorResponse | undefined) {
    if (err?.status === "KO") {
        return { errored: true, message: err.message };
    }

    return { errored: false, message: null };
}

export function hasServiceError<T>(resp: AxiosResponse<ServiceResponse<T>>) {
    const errResp = toErrorResponse(resp);

    return hasError(errResp);
}

export function validateStatusAuthentication(status: number) {
    return (status >= 200 && status < 300) || status === 401 || status === 403;
}

export function validateStatusOneTimeCode(status: number) {
    return status === 401 || status === 403 || (status >= 200 && status < 400);
}

export function validateStatusWebAuthnCreation(status: number) {
    return status < 300 || status === 409;
}
