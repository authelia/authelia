import { FirstFactorPath, FirstFactorReauthenticatePath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface PostFirstFactorBody {
    username: string;
    password: string;
    keepMeLoggedIn: boolean;
    targetURL?: string;
    requestMethod?: string;
    workflow?: string;
    workflowID?: string;
}

interface PostFirstFactorReauthenticateBody {
    password: string;
    targetURL?: string;
    requestMethod?: string;
    workflow?: string;
    workflowID?: string;
}

export async function postFirstFactor(
    username: string,
    password: string,
    rememberMe: boolean,
    targetURL?: string,
    requestMethod?: string,
    workflow?: string,
    workflowID?: string,
) {
    const data: PostFirstFactorBody = {
        username,
        password,
        keepMeLoggedIn: rememberMe,
    };

    if (targetURL) {
        data.targetURL = targetURL;
    }

    if (requestMethod) {
        data.requestMethod = requestMethod;
    }

    if (workflow) {
        data.workflow = workflow;
    }

    if (workflowID) {
        data.workflowID = workflowID;
    }

    const res = await PostWithOptionalResponse<SignInResponse>(FirstFactorPath, data);
    return res ? res : ({} as SignInResponse);
}

export async function postFirstFactorReauthenticate(
    password: string,
    targetURL?: string,
    requestMethod?: string,
    workflow?: string,
    workflowID?: string,
) {
    const data: PostFirstFactorReauthenticateBody = {
        password,
    };

    if (targetURL) {
        data.targetURL = targetURL;
    }

    if (requestMethod) {
        data.requestMethod = requestMethod;
    }

    if (workflow) {
        data.workflow = workflow;
    }

    if (workflowID) {
        data.workflowID = workflowID;
    }

    const res = await PostWithOptionalResponse<SignInResponse>(FirstFactorReauthenticatePath, data);
    return res ? res : ({} as SignInResponse);
}
