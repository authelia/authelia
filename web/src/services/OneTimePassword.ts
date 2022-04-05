import { toWorkflowPath, Workflow } from "@models/Workflow";
import { CompleteTOTPSignInPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface CompleteTOTPSigninBody {
    token: string;
    targetURL?: string;
}

export function completeTOTPSignIn(passcode: string, targetURL: string | undefined, workflow: Workflow) {
    const body: CompleteTOTPSigninBody = { token: `${passcode}` };
    if (targetURL) {
        body.targetURL = targetURL;
    }
    return PostWithOptionalResponse<SignInResponse>(toWorkflowPath(CompleteTOTPSignInPath, workflow), body);
}
