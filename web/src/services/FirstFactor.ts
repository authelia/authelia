import { FirstFactorPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface PostFirstFactorBody {
    username: string;
    password: string;
    keepMeLoggedIn: boolean;
    targetURL?: string;
    requestMethod?: string;
    workflow?: string;
}

export async function postFirstFactor(
    username: string,
    password: string,
    rememberMe: boolean,
    targetURL: string | null,
    requestMethod: string | null,
    workflow: string | null,
) {
    const data: PostFirstFactorBody = {
        username: username,
        password: password,
        keepMeLoggedIn: rememberMe,
        targetURL: targetURL === null ? undefined : targetURL,
        requestMethod: requestMethod === null ? undefined : requestMethod,
        workflow: workflow == null ? undefined : workflow,
    };

    const res = await PostWithOptionalResponse<SignInResponse>(FirstFactorPath, data);
    return res ? res : ({} as SignInResponse);
}
