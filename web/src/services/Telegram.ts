import { CompleteTelegramSignInPath, TelegramInitPath, TelegramStatusPath } from "@services/Api";
import { Get, PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

export interface TelegramInitResponse {
    token: string;
    bot_username: string;
    bot_deep_link: string;
}

export interface TelegramStatusResponse {
    verified: boolean;
    expired: boolean;
}

interface CompleteTelegramSignInBody {
    token: string;
    targetURL?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
    userCode?: string;
}

export async function initiateTelegramSignIn() {
    return Get<TelegramInitResponse>(TelegramInitPath);
}

export async function getTelegramStatus(token: string) {
    return Get<TelegramStatusResponse>(`${TelegramStatusPath}/${token}`);
}

export async function completeTelegramSignIn(
    token: string,
    targetURL?: string,
    flowID?: string,
    flow?: string,
    subflow?: string,
    userCode?: string,
) {
    const body: CompleteTelegramSignInBody = {
        flow,
        flowID,
        subflow,
        targetURL,
        token,
        userCode,
    };

    return PostWithOptionalResponse<SignInResponse>(CompleteTelegramSignInPath, body);
}
