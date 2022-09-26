import { Workflow } from "@hooks/Workflow";
import { CompleteTOTPSignInPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface CompleteTOTPSigninBody {
    token: string;
    targetURL?: string;
    workflow?: string;
    workflowID?: string;
}

export function completeTOTPSignIn(passcode: string, targetURL?: string, workflow?: Workflow) {
    const body: CompleteTOTPSigninBody = { token: `${passcode}` };
    if (targetURL) {
        body.targetURL = targetURL;
    }

    if (workflow) {
        body.workflow = workflow.name;
        body.workflowID = workflow.id;
    }

    return PostWithOptionalResponse<SignInResponse>(CompleteTOTPSignInPath, body);
}
