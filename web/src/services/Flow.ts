// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { FlowContinuePath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

export async function postFlowContinue(
    flowID: string | undefined,
    flow: string | undefined,
    subflow: string | undefined,
    userCode: string | undefined,
) {
    const data = {
        flow,
        flowID,
        subflow,
        userCode,
    };

    const res = await PostWithOptionalResponse<SignInResponse>(FlowContinuePath, data);

    return res ?? ({} as SignInResponse);
}
