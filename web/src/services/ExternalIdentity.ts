// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import i18n from "i18next";

import {
    ExternalIdentityLinks,
    ExternalIdentityProvider,
    ExternalIdentityStartResponse,
} from "@models/ExternalIdentity";
import {
    FirstFactorExternalIdentityPath,
    UserExternalIdentityLinkPath,
    UserExternalIdentityLinkPendingPath,
    UserExternalIdentityLinksPath,
} from "@services/Api";
import { DeleteWithOptionalResponse, Get, Post, PutWithOptionalResponse } from "@services/Client";

export async function getExternalIdentityProviders(signal?: AbortSignal): Promise<ExternalIdentityProvider[]> {
    const res = await Get<{ providers: ExternalIdentityProvider[] }>(FirstFactorExternalIdentityPath, signal);

    return res.providers;
}

export interface PostExternalIdentityStartBody {
    targetURL?: string;
    requestMethod?: string;
    keepMeLoggedIn: boolean;
}

export async function postExternalIdentityStart(
    id: string,
    body: PostExternalIdentityStartBody,
    signal?: AbortSignal,
): Promise<ExternalIdentityStartResponse> {
    return Post<ExternalIdentityStartResponse>(
        `${FirstFactorExternalIdentityPath}/${encodeURIComponent(id)}`,
        { ...body, language: i18n.resolvedLanguage || i18n.language },
        signal,
    );
}

export async function getExternalIdentityLinks(signal?: AbortSignal): Promise<ExternalIdentityLinks> {
    return Get<ExternalIdentityLinks>(UserExternalIdentityLinksPath, signal);
}

export async function putExternalIdentityLink(signal?: AbortSignal): Promise<void> {
    await PutWithOptionalResponse<void>(UserExternalIdentityLinkPath, {}, signal);
}

export async function deleteExternalIdentityLinkPending(signal?: AbortSignal): Promise<void> {
    await DeleteWithOptionalResponse<void>(UserExternalIdentityLinkPendingPath, undefined, signal);
}

export async function deleteExternalIdentityLink(id: number, signal?: AbortSignal): Promise<void> {
    await DeleteWithOptionalResponse<void>(`${UserExternalIdentityLinkPath}/${id}`, undefined, signal);
}
