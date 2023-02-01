import { CompleteTOTPRegistrationPath, InitiateTOTPRegistrationPath } from "@services/Api";
import { Post, PostWithOptionalResponse } from "@services/Client";

export async function initiateTOTPRegistrationProcess() {
    await PostWithOptionalResponse(InitiateTOTPRegistrationPath);
}

interface CompleteTOTPRegistrationResponse {
    base32_secret: string;
    otpauth_url: string;
}

export async function completeTOTPRegistrationProcess(processToken: string) {
    return Post<CompleteTOTPRegistrationResponse>(CompleteTOTPRegistrationPath, { token: processToken });
}
