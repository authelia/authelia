import axios from "axios";

import {
    ErrorResponse,
    OKResponse,
    ServiceResponse,
    UserSessionElevationPath,
    hasServiceError,
    toData,
    validateStatusOneTimeCode,
} from "@services/Api";

export interface UserSessionElevation {
    require_second_factor: boolean;
    skip_second_factor: boolean;
    can_skip_second_factor: boolean;
    factor_knowledge: boolean;
    elevated: boolean;
    expires: number;
}

export interface UserSessionElevationGenerateData {
    delete_id: string;
}

export async function getUserSessionElevation() {
    const res = await axios<ServiceResponse<UserSessionElevation>>({
        method: "GET",
        url: UserSessionElevationPath,
    });

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new Error(
            `Failed GET to ${UserSessionElevationPath}. Code: ${res.status}. Message: ${hasServiceError(res).message}`,
        );
    }

    return toData<UserSessionElevation>(res);
}

export async function generateUserSessionElevation() {
    const res = await axios<ServiceResponse<UserSessionElevationGenerateData>>({
        method: "POST",
        url: UserSessionElevationPath,
    });

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new Error(
            `Failed POST to ${UserSessionElevationPath}. Code: ${res.status}. Message: ${hasServiceError(res).message}`,
        );
    }

    return toData<UserSessionElevationGenerateData>(res);
}

export async function verifyUserSessionElevation(otc: string) {
    const res = await axios<OKResponse | ErrorResponse>({
        method: "PUT",
        url: UserSessionElevationPath,
        data: { otc: otc },
        validateStatus: validateStatusOneTimeCode,
    });

    return res.status === 200 && res.data.status === "OK";
}

export async function deleteUserSessionElevation(deleteID: string) {
    const res = await axios<OKResponse | ErrorResponse>({
        method: "DELETE",
        url: `${UserSessionElevationPath}/${deleteID}`,
    });

    return res.status === 200 && res.data.status === "OK";
}
