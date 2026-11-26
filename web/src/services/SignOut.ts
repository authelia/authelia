// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { LogoutPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";

export type SignOutResponse = { safeTargetURL: boolean } | undefined;

export type SignOutBody = {
    targetURL?: string;
    flowID?: string;
};

export async function signOut(
    targetURL: string | undefined,
    signal?: AbortSignal,
    flowID?: string,
): Promise<SignOutResponse> {
    const body: SignOutBody = {};
    if (targetURL) {
        body.targetURL = targetURL;
    }
    if (flowID) {
        body.flowID = flowID;
    }

    return PostWithOptionalResponse<SignOutResponse>(LogoutPath, body, signal);
}
