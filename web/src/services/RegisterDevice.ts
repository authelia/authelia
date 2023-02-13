import { CompleteTOTPRegistrationPath, InitiateTOTPRegistrationPath, TOTPRegistrationPath } from "@services/Api";
import { Post, PostWithOptionalResponse, Put } from "@services/Client";

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

export async function getTOTPSecret(algorithm: string, length: number, period: number) {
    return Put<CompleteTOTPRegistrationResponse>(TOTPRegistrationPath, {
        algorithm: algorithm,
        length: length,
        period: period,
    });
}
