import { AxiosResponse } from "axios";

import { getBasePath } from "@utils/BasePath";

const basePath = getBasePath();

export const OpenIDConnectConsentPath = basePath + "/api/oidc/consent";
export const OpenIDConnectDeviceAuthorizationPath = basePath + "/api/oidc/device-authorization";

export const FirstFactorPath = basePath + "/api/firstfactor";
export const FirstFactorPasskeyPath = basePath + "/api/firstfactor/passkey";
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

export const AdminConfigPath = basePath + "/api/admin/config";
export const AdminUserRestPath = basePath + "/api/admin/users";
export const AdminUserFieldMetadataPath = basePath + "/api/admin/user-fields";

export const LocaleInformationPath = basePath + "/locales";

export interface AuthenticationErrorResponse extends ErrorResponse {
    authentication: boolean;
    elevation: boolean;
}

export interface ErrorResponse {
    status: "KO";
    message: string;
}

export interface ErrorRateLimitResponse extends ErrorResponse {
    limited: boolean;
    retryAfter: number;
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

export type RateLimitedData<T> = {
    data?: T;
    limited: boolean;
    retryAfter: number;
};

function getRetryAfter(resp: AxiosResponse): number {
    if (resp.status !== 429) {
        return 0;
    }

    const valueRetryAfter = resp.headers["retry-after"];
    if (valueRetryAfter) {
        if (/^\d+$/.test(valueRetryAfter)) {
            const retryAfter = Number.parseFloat(valueRetryAfter);

            if (Number.isNaN(retryAfter)) {
                throw new TypeError("Header Retry-After has an invalid number value");
            }

            return retryAfter;
        } else {
            const date = new Date(valueRetryAfter);
            if (Number.isNaN(date.getTime())) {
                throw new TypeError("Header Retry-After has an invalid date value");
            }

            return Math.max(0, (date.getTime() - Date.now()) / 1000);
        }
    }

    throw new Error("Header Retry-After is missing");
}

export function toDataRateLimited<T>(resp: AxiosResponse<ServiceResponse<T>>): RateLimitedData<T> | undefined {
    if (resp.data && "status" in resp.data) {
        if (resp.data["status"] === "OK") {
            return { data: resp.data.data, limited: false, retryAfter: 0 };
        } else if (resp.data["status"] === "KO") {
            return { limited: resp.status === 429, retryAfter: getRetryAfter(resp) };
        } else if (resp.status === 429) {
            return { limited: true, retryAfter: getRetryAfter(resp) };
        }
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

export function validateStatusTooManyRequests(status: number) {
    return (status >= 200 && status < 300) || status === 429;
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
