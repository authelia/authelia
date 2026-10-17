// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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
    const res = await axios<ErrorResponse | OKResponse>({
        data: { token: token },
        method: "DELETE",
        url: ResetPasswordPath,
    });

    return { ok: res.status === 200 && res.data.status === "OK", status: res.status };
}
