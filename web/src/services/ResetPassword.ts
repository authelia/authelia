import axios from "axios";

import {
    CompleteResetPasswordPath,
    ErrorResponse,
    InitiateResetPasswordPath,
    OKResponse,
    ResetPasswordPath,
} from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";

export async function initiateResetPasswordProcess(username: string) {
    return PostWithOptionalResponse(InitiateResetPasswordPath, { username });
}

export async function completeResetPasswordProcess(token: string) {
    return PostWithOptionalResponse(CompleteResetPasswordPath, { token });
}

export async function resetPassword(newPassword: string) {
    return PostWithOptionalResponse(ResetPasswordPath, { password: newPassword });
}

export async function deleteResetPasswordToken(token: string) {
    const res = await axios<OKResponse | ErrorResponse>({
        method: "DELETE",
        url: ResetPasswordPath,
        data: { token: token },
    });

    return res.status === 200 && res.data.status === "OK";
}
