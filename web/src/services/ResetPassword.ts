import axios from "axios";

import {
    ErrorResponse,
    IdentityVerification,
    InitiateResetPasswordIdentityVerificationPath,
    OKResponse,
    ResetPasswordPath,
} from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";

export async function initiateResetPasswordIVProcess(username: string) {
    return PostWithOptionalResponse(InitiateResetPasswordIdentityVerificationPath, { username });
}

export async function completeResetPasswordIVProcess(token: string) {
    return PostWithOptionalResponse(IdentityVerification, { token });
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
