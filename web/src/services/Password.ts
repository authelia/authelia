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
    userCode?: string;
}

interface PostFirstFactorReauthenticateBody {
    password: string;
    targetURL?: string;
    requestMethod?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
    userCode?: string;
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
    userCode?: string,
) {
    const data: PostFirstFactorBody = {
        flow,
        flowID,
        keepMeLoggedIn: rememberMe,
        password,
        requestMethod,
        subflow,
        targetURL,
        userCode,
        username,
    };

    const res = await PostWithOptionalResponse<SignInResponse>(FirstFactorPath, data);
    return res ?? ({} as SignInResponse);
}

export async function postFirstFactorReauthenticate(
    password: string,
    targetURL?: string,
    requestMethod?: string,
    flowID?: string,
    flow?: string,
    subflow?: string,
    userCode?: string,
) {
    const data: PostFirstFactorReauthenticateBody = {
        flow,
        flowID,
        password,
        requestMethod,
        subflow,
        targetURL,
        userCode,
    };

    const res = await PostWithOptionalResponse<SignInResponse>(FirstFactorReauthenticatePath, data);
    return res ?? ({} as SignInResponse);
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
        flow,
        flowID,
        password,
        subflow,
        targetURL,
    };

    const res = await PostWithOptionalResponse<SignInResponse>(CompletePasswordSignInPath, data);
    return res ?? ({} as SignInResponse);
}
