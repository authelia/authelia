import { TOTPRegistrationPath } from "@services/Api";
import { Put } from "@services/Client";

interface CompleteTOTPRegistrationResponse {
    base32_secret: string;
    otpauth_url: string;
}

export async function getTOTPSecret(algorithm: string, length: number, period: number) {
    return Put<CompleteTOTPRegistrationResponse>(TOTPRegistrationPath, {
        algorithm: algorithm,
        length: length,
        period: period,
    });
}
