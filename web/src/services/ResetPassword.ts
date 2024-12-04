import axios from "axios";

import {
    CompleteResetPasswordPath,
    ErrorResponse,
    InitiateResetPasswordPath,
    OKResponse,
    ResetPasswordPath,
    ServiceResponse,
    hasServiceError,
    validateStatusTooManyRequests,
} from "@services/Api";

async function postWithPossibleTooManyRequests(path: string, body?: any) {
    const res = await axios.post<ServiceResponse<undefined>>(path, body, {
        validateStatus: validateStatusTooManyRequests,
    });

    if (res.status === 429) {
        return false;
    }

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new Error(`Failed POST to ${path}. Code: ${res.status}. Message: ${hasServiceError(res).message}`);
    }

    return true;
}

export async function initiateResetPasswordProcess(username: string) {
    return postWithPossibleTooManyRequests(InitiateResetPasswordPath, { username });
}

export async function completeResetPasswordProcess(token: string) {
    return postWithPossibleTooManyRequests(CompleteResetPasswordPath, { token });
}

export async function resetPassword(newPassword: string) {
    return postWithPossibleTooManyRequests(ResetPasswordPath, { password: newPassword });
}

export async function deleteResetPasswordToken(token: string) {
    const res = await axios<OKResponse | ErrorResponse>({
        method: "DELETE",
        url: ResetPasswordPath,
        data: { token: token },
        validateStatus: validateStatusTooManyRequests,
    });

    return { ok: res.status === 200 && res.data.status === "OK", status: res.status };
}
