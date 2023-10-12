import axios from "axios";

import {
    TOTPAlgorithmPayload,
    TOTPDigits,
    TOTPOptions,
    UserInfoTOTPConfiguration,
    toEnum,
} from "@models/TOTPConfiguration";
import {
    AuthenticationOKResponse,
    CompleteTOTPSignInPath,
    ServiceResponse,
    TOTPConfigurationPath,
    TOTPRegistrationPath,
    validateStatusAuthentication,
} from "@services/Api";
import { Get } from "@services/Client";

export interface UserInfoTOTPConfigurationPayload {
    created_at: string;
    last_used_at?: string;
    issuer: string;
    algorithm: TOTPAlgorithmPayload;
    digits: TOTPDigits;
    period: number;
}

function toUserInfoTOTPConfiguration(payload: UserInfoTOTPConfigurationPayload): UserInfoTOTPConfiguration {
    return {
        created_at: new Date(payload.created_at),
        last_used_at: payload.last_used_at ? new Date(payload.last_used_at) : undefined,
        issuer: payload.issuer,
        algorithm: toEnum(payload.algorithm),
        digits: payload.digits,
        period: payload.period,
    };
}

export async function getUserInfoTOTPConfiguration(): Promise<UserInfoTOTPConfiguration> {
    const res = await Get<UserInfoTOTPConfigurationPayload>(TOTPConfigurationPath);

    return toUserInfoTOTPConfiguration(res);
}

export async function getUserInfoTOTPConfigurationOptional(): Promise<UserInfoTOTPConfiguration | null> {
    const res = await axios.get<ServiceResponse<UserInfoTOTPConfigurationPayload>>(TOTPConfigurationPath, {
        validateStatus: function (status) {
            return status < 300 || status === 404;
        },
    });

    if (res === null || res.status === 404 || res.data.status === "KO") {
        return null;
    }

    return toUserInfoTOTPConfiguration(res.data.data);
}

export interface TOTPOptionsPayload {
    algorithm: TOTPAlgorithmPayload;
    algorithms: TOTPAlgorithmPayload[];
    length: TOTPDigits;
    lengths: TOTPDigits[];
    period: number;
    periods: number[];
}

export async function getTOTPOptions(): Promise<TOTPOptions> {
    const res = await Get<TOTPOptionsPayload>(TOTPRegistrationPath);

    return {
        algorithm: toEnum(res.algorithm),
        algorithms: res.algorithms.map((alg) => toEnum(alg)),
        length: res.length,
        lengths: res.lengths,
        period: res.period,
        periods: res.periods,
    };
}

export async function deleteUserTOTPConfiguration() {
    return await axios<AuthenticationOKResponse>({
        method: "DELETE",
        url: CompleteTOTPSignInPath,
        validateStatus: validateStatusAuthentication,
    });
}
