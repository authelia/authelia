import { AxiosResponse } from "axios";

export const FirstFactorPath = "/auth/api/firstfactor";
export const InitiateTOTPRegistrationPath = "/auth/api/secondfactor/totp/identity/start";
export const CompleteTOTPRegistrationPath = "/auth/api/secondfactor/totp/identity/finish";

export const InitiateU2FRegistrationPath = "/auth/api/secondfactor/u2f/identity/start";
export const CompleteU2FRegistrationStep1Path = "/auth/api/secondfactor/u2f/identity/finish";
export const CompleteU2FRegistrationStep2Path = "/auth/api/secondfactor/u2f/register";

export const InitiateU2FSignInPath = "/auth/api/secondfactor/u2f/sign_request";
export const CompleteU2FSignInPath = "/auth/api/secondfactor/u2f/sign";

export const CompletePushNotificationSignInPath = "/auth/api/secondfactor/duo"
export const CompleteTOTPSignInPath = "/auth/api/secondfactor/totp"

export const InitiateResetPasswordPath = "/auth/api/reset-password/identity/start";
export const CompleteResetPasswordPath = "/auth/api/reset-password/identity/finish";
// Do the password reset during completion.
export const ResetPasswordPath = "/auth/api/reset-password"

export const LogoutPath = "/auth/api/logout";
export const StatePath = "/auth/api/state";
export const UserInfoPath = "/auth/api/user/info";
export const UserInfo2FAMethodPath = "/auth/api/user/info/2fa_method";

export const ConfigurationPath = "/auth/api/configuration";
export const ExtendedConfigurationPath = "/auth/api/configuration/extended";

export interface ErrorResponse {
    status: "KO";
    message: string;
}

export interface Response<T> {
    status: "OK";
    data: T;
}

export type ServiceResponse<T> = Response<T> | ErrorResponse;

function toErrorResponse<T>(resp: AxiosResponse<ServiceResponse<T>>): ErrorResponse | undefined {
    if (resp.data && "status" in resp.data && resp.data["status"] === "KO") {
        return resp.data as ErrorResponse;
    }
    return undefined;
}

export function toData<T>(resp: AxiosResponse<ServiceResponse<T>>): T | undefined {
    if (resp.data && "status" in resp.data && resp.data["status"] === "OK") {
        return resp.data.data as T;
    }
    return undefined
}

export function hasServiceError<T>(resp: AxiosResponse<ServiceResponse<T>>) {
    const errResp = toErrorResponse(resp);
    if (errResp && errResp.status === "KO") {
        return true;
    }
    return false;
}