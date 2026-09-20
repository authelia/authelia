// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { GenerateRecoveryCodesResponse, RecoveryCodesStatus, RecoveryCodesStatusPayload } from "@models/RecoveryCodes";
import { CompleteRecoveryCodeSignInPath, RecoveryCodesGeneratePath, RecoveryCodesStatusPath } from "@services/Api";
import { Get, Post, PostWithOptionalResponse } from "@services/Client";

function toRecoveryCodesStatus(payload: RecoveryCodesStatusPayload): RecoveryCodesStatus {
    return {
        codes_remaining: payload.codes_remaining,
        codes_total: payload.codes_total,
        generated_at: payload.generated_at ? new Date(payload.generated_at) : undefined,
        last_used_at: payload.last_used_at ? new Date(payload.last_used_at) : undefined,
    };
}

export async function getRecoveryCodesStatus(): Promise<RecoveryCodesStatus> {
    const payload = await Get<RecoveryCodesStatusPayload>(RecoveryCodesStatusPath);

    return toRecoveryCodesStatus(payload);
}

export async function generateRecoveryCodes(): Promise<GenerateRecoveryCodesResponse> {
    return Post<GenerateRecoveryCodesResponse>(RecoveryCodesGeneratePath);
}

export interface SignInRecoveryCodeParams {
    code: string;
    targetURL?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
    userCode?: string;
}

export async function signInWithRecoveryCode(params: SignInRecoveryCodeParams): Promise<void> {
    await PostWithOptionalResponse(CompleteRecoveryCodeSignInPath, {
        code: params.code,
        flow: params.flow,
        flowID: params.flowID,
        subflow: params.subflow,
        targetURL: params.targetURL,
        userCode: params.userCode,
    });
}
