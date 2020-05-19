import { AxiosResponse } from "axios";

export const FirstFactorPath = (window as any).Base + "/api/firstfactor";
export const InitiateTOTPRegistrationPath = (window as any).Base + "/api/secondfactor/totp/identity/start";
export const CompleteTOTPRegistrationPath = (window as any).Base + "/api/secondfactor/totp/identity/finish";

export const InitiateU2FRegistrationPath = (window as any).Base + "/api/secondfactor/u2f/identity/start";
export const CompleteU2FRegistrationStep1Path = (window as any).Base + "/api/secondfactor/u2f/identity/finish";
export const CompleteU2FRegistrationStep2Path = (window as any).Base + "/api/secondfactor/u2f/register";

export const InitiateU2FSignInPath = (window as any).Base + "/api/secondfactor/u2f/sign_request";
export const CompleteU2FSignInPath = (window as any).Base + "/api/secondfactor/u2f/sign";

export const CompletePushNotificationSignInPath = (window as any).Base + "/api/secondfactor/duo"
export const CompleteTOTPSignInPath = (window as any).Base + "/api/secondfactor/totp"

export const InitiateResetPasswordPath = (window as any).Base + "/api/reset-password/identity/start";
export const CompleteResetPasswordPath = (window as any).Base + "/api/reset-password/identity/finish";
// Do the password reset during completion.
export const ResetPasswordPath = (window as any).Base + "/api/reset-password"

export const LogoutPath = (window as any).Base + "/api/logout";
export const StatePath = (window as any).Base + "/api/state";
export const UserInfoPath = (window as any).Base + "/api/user/info";
export const UserInfo2FAMethodPath = (window as any).Base + "/api/user/info/2fa_method";

export const ConfigurationPath = (window as any).Base + "/api/configuration";
export const ExtendedConfigurationPath = (window as any).Base + "/api/configuration/extended";

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