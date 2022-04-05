import { toWorkflowPath, Workflow } from "@models/Workflow";
import { FirstFactorPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface PostFirstFactorBody {
    username: string;
    password: string;
    keepMeLoggedIn: boolean;
    targetURL?: string;
    requestMethod?: string;
}

export async function postFirstFactor(
    username: string,
    password: string,
    rememberMe: boolean,
    workflow: Workflow,
    targetURL?: string,
    requestMethod?: string,
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

    const res = await PostWithOptionalResponse<SignInResponse>(toWorkflowPath(FirstFactorPath, workflow), data);
    return res ? res : ({} as SignInResponse);
}
