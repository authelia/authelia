import { Workflow } from "@hooks/Workflow";
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
    workflowID?: string;
}

export async function postFirstFactor(
    username: string,
    password: string,
    rememberMe: boolean,
    targetURL?: string,
    requestMethod?: string,
    workflow?: Workflow,
) {
    const body: PostFirstFactorBody = {
        username,
        password,
        keepMeLoggedIn: rememberMe,
    };

    if (targetURL) {
        body.targetURL = targetURL;
    }

    if (requestMethod) {
        body.requestMethod = requestMethod;
    }

    if (workflow) {
        body.workflow = workflow.name;
        body.workflowID = workflow.id;
    }

    const res = await PostWithOptionalResponse<SignInResponse>(FirstFactorPath, body);
    return res ? res : ({} as SignInResponse);
}
