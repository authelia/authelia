import { CompletePasswordSignInPath, FirstFactorPath } from "@services/Api";
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
    targetURL?: string,
    requestMethod?: string,
    workflow?: string,
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

    const res = await PostWithOptionalResponse<SignInResponse>(FirstFactorPath, data);
    return res ? res : ({} as SignInResponse);
}

interface PostSecondFactorBody {
    password: string;
    targetURL?: string;
    workflow?: string;
    workflowID?: string;
}

export async function postSecondFactor(
    password: string,
    targetURL?: string | undefined,
    workflow?: string,
    workflowID?: string,
) {
    const data: PostSecondFactorBody = {
        password,
    };

    if (targetURL) {
        data.targetURL = targetURL;
    }

    if (workflow) {
        data.workflow = workflow;
    }

    if (workflowID) {
        data.workflowID = workflowID;
    }

    const res = await PostWithOptionalResponse<SignInResponse>(CompletePasswordSignInPath, data);
    return res ? res : ({} as SignInResponse);
}
