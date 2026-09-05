import {
    OpenIDConnectLinks,
    OpenIDConnectProvider,
    OpenIDConnectStartResponse,
} from "@models/OpenIDConnectRelyingParty";
import {
    FirstFactorOpenIDConnectPath,
    UserOpenIDConnectLinkPath,
    UserOpenIDConnectLinkPendingPath,
    UserOpenIDConnectLinksPath,
} from "@services/Api";
import { DeleteWithOptionalResponse, Get, Post, PutWithOptionalResponse } from "@services/Client";

export async function getOpenIDConnectProviders(signal?: AbortSignal): Promise<OpenIDConnectProvider[]> {
    const res = await Get<{ providers: OpenIDConnectProvider[] }>(FirstFactorOpenIDConnectPath, signal);

    return res.providers;
}

export interface PostOpenIDConnectStartBody {
    targetURL?: string;
    requestMethod?: string;
    keepMeLoggedIn: boolean;
}

export async function postOpenIDConnectStart(
    id: string,
    body: PostOpenIDConnectStartBody,
    signal?: AbortSignal,
): Promise<OpenIDConnectStartResponse> {
    return Post<OpenIDConnectStartResponse>(`${FirstFactorOpenIDConnectPath}/${encodeURIComponent(id)}`, body, signal);
}

export async function getOpenIDConnectLinks(signal?: AbortSignal): Promise<OpenIDConnectLinks> {
    return Get<OpenIDConnectLinks>(UserOpenIDConnectLinksPath, signal);
}

export async function putOpenIDConnectLink(signal?: AbortSignal): Promise<void> {
    await PutWithOptionalResponse<void>(UserOpenIDConnectLinkPath, {}, signal);
}

export async function deleteOpenIDConnectLinkPending(signal?: AbortSignal): Promise<void> {
    await DeleteWithOptionalResponse<void>(UserOpenIDConnectLinkPendingPath, undefined, signal);
}

export async function deleteOpenIDConnectLink(id: number, signal?: AbortSignal): Promise<void> {
    await DeleteWithOptionalResponse<void>(`${UserOpenIDConnectLinkPath}/${id}`, undefined, signal);
}
