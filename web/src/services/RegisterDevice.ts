import U2fApi from "u2f-api";

import {
    InitiateTOTPRegistrationPath,
    CompleteTOTPRegistrationPath,
    InitiateU2FRegistrationPath,
    CompleteU2FRegistrationStep1Path,
    CompleteU2FRegistrationStep2Path,
} from "./Api";
import { Post, PostWithOptionalResponse } from "./Client";

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

export async function initiateU2FRegistrationProcess() {
    return PostWithOptionalResponse(InitiateU2FRegistrationPath);
}

interface U2RRegistrationStep1Response {
    appId: string;
    registerRequests: [
        {
            version: string;
            challenge: string;
        },
    ];
}

export async function completeU2FRegistrationProcessStep1(processToken: string) {
    return Post<U2RRegistrationStep1Response>(CompleteU2FRegistrationStep1Path, { token: processToken });
}

export async function completeU2FRegistrationProcessStep2(response: U2fApi.RegisterResponse) {
    return PostWithOptionalResponse(CompleteU2FRegistrationStep2Path, response);
}
