import axios from "axios";

import { OKResponse, ServiceResponse, UserSessionElevationPath, hasServiceError, toData } from "@services/Api";

export async function getElevation() {
    var res = await axios<ServiceResponse<SessionElevationData>>({
        method: "GET",
        url: UserSessionElevationPath,
    });

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new Error(
            `Failed POST to ${UserSessionElevationPath}. Code: ${res.status}. Message: ${hasServiceError(res).message}`,
        );
    }

    return toData<SessionElevationData>(res);
}

export interface SessionElevationData {
    elevated: boolean;
    expires: number;
    id: number;
    remote_ip: string;
    public_id: string;
    delete_id: string;
    signature: string;
}

export interface SessionElevationStartData {
    expires: number;
    public_id: string;
    delete_id: string;
    signature: string;
    password: string;
}

export async function startElevation() {
    var res = await axios<ServiceResponse<SessionElevationStartData>>({
        method: "POST",
        url: UserSessionElevationPath,
    });

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new Error(
            `Failed POST to ${UserSessionElevationPath}. Code: ${res.status}. Message: ${hasServiceError(res).message}`,
        );
    }

    return toData<SessionElevationStartData>(res);
}

export async function finishElevation(otp: string) {
    return axios<OKResponse>({
        method: "PUT",
        url: UserSessionElevationPath,
        data: { otp: otp },
    });
}

export async function deleteElevation(deleteID: string) {
    return axios<OKResponse>({
        method: "DELETE",
        url: `${UserSessionElevationPath}/${deleteID}`,
    });
}
