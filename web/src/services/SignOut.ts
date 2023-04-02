// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { LogoutPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";

export type SignOutResponse = { safeTargetURL: boolean } | undefined;

export type SignOutBody = {
    targetURL?: string;
};

export async function signOut(targetURL: string | undefined): Promise<SignOutResponse> {
    const body: SignOutBody = {};
    if (targetURL) {
        body.targetURL = targetURL;
    }

    return PostWithOptionalResponse<SignOutResponse>(LogoutPath, body);
}
