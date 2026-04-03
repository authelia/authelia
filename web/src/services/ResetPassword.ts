import axios from "axios";

import {
    CompleteResetPasswordPath,
    ErrorResponse,
    InitiateResetPasswordPath,
    OKResponse,
    ResetPasswordPath,
    validateStatusTooManyRequests,
} from "@services/Api";
import { PostWithOptionalResponse, PostWithOptionalResponseRateLimited } from "@services/Client";

export async function initiateResetPasswordProcess(username: string) {
    return PostWithOptionalResponseRateLimited(InitiateResetPasswordPath, { username });
}

export async function completeResetPasswordProcess(token: string) {
    return PostWithOptionalResponseRateLimited(CompleteResetPasswordPath, { token });
}

export async function resetPassword(newPassword: string) {
    return PostWithOptionalResponse(ResetPasswordPath, { password: newPassword });
}

export async function deleteResetPasswordToken(token: string) {
    const res = await axios<ErrorResponse | OKResponse>({
        data: { token: token },
        method: "DELETE",
        url: ResetPasswordPath,
        validateStatus: validateStatusTooManyRequests,
    });

    return { ok: res.status === 200 && res.data.status === "OK", status: res.status };
}
