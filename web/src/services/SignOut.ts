// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { FlowID } from "@constants/SearchParams";
import { LogoutPath } from "@services/Api";
import { DeleteWithOptionalResponse, Get, PostWithOptionalResponse } from "@services/Client";

export type SignOutResponse = { redirectURL?: string; safeTargetURL: boolean } | undefined;

export type SignOutBody = {
    flowID?: string;
    targetURL?: string;
};

export type SignOutPending = {
    clientID?: string;
    clientName?: string;
    pending: boolean;
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

export async function getSignOutPending(flowID: string, signal?: AbortSignal): Promise<SignOutPending> {
    return Get<SignOutPending>(flowPath(flowID), signal);
}

export async function cancelSignOut(flowID: string, signal?: AbortSignal) {
    return DeleteWithOptionalResponse(flowPath(flowID), undefined, signal);
}

function flowPath(flowID: string) {
    return `${LogoutPath}?${new URLSearchParams({ [FlowID]: flowID }).toString()}`;
}
