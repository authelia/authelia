import { CompleteResetPasswordPath, InitiateResetPasswordPath, ResetPasswordPath } from "@services/Api";
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
