import { ChecksSafeRedirectionPath, CompleteResetPasswordPath, ResetPasswordPath } from "./Api";
import { PostWithOptionalResponse } from "./Client";

interface SafeRedirectionResponse {
    ok: boolean;
}

export async function checkSafeRedirection(uri: string) {
    return PostWithOptionalResponse<SafeRedirectionResponse>(ChecksSafeRedirectionPath, { uri });
}
