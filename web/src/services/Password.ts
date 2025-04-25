import { CompletePasswordSignInPath, FirstFactorPath, FirstFactorReauthenticatePath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface PostFirstFactorBody {
    username: string;
    password: string;
    keepMeLoggedIn: boolean;
    targetURL?: string;
    requestMethod?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
}

interface PostFirstFactorReauthenticateBody {
    password: string;
    targetURL?: string;
    requestMethod?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
}

export async function postFirstFactor(
    username: string,
    password: string,
    rememberMe: boolean,
    targetURL?: string,
    requestMethod?: string,
    flowID?: string,
    flow?: string,
    subflow?: string,
) {
    const data: PostFirstFactorBody = {
        username,
        password,
        keepMeLoggedIn: rememberMe,
        targetURL,
        requestMethod,
        flowID,
        flow,
        subflow,
    };

    const res = await PostWithOptionalResponse<SignInResponse>(FirstFactorPath, data);
    return res ? res : ({} as SignInResponse);
}

export async function postFirstFactorReauthenticate(
    password: string,
    targetURL?: string,
    requestMethod?: string,
    flowID?: string,
    flow?: string,
    subflow?: string,
) {
    const data: PostFirstFactorReauthenticateBody = {
        password,
        targetURL,
        requestMethod,
        flowID,
        flow,
        subflow,
    };

    const res = await PostWithOptionalResponse<SignInResponse>(FirstFactorReauthenticatePath, data);
    return res ? res : ({} as SignInResponse);
}

interface PostSecondFactorBody {
    password: string;
    targetURL?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
}

export async function postSecondFactor(
    password: string,
    targetURL?: string | undefined,
    flowID?: string,
    flow?: string,
    subflow?: string,
) {
    const data: PostSecondFactorBody = {
        password,
        targetURL,
        flowID,
        flow,
        subflow,
    };

    const res = await PostWithOptionalResponse<SignInResponse>(CompletePasswordSignInPath, data);
    return res ? res : ({} as SignInResponse);
}
