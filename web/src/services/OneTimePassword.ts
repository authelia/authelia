import { CompleteTOTPSignInPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface CompleteTOTPSigninBody {
    token: string;
    targetURL?: string;
    workflow?: string;
}

export function completeTOTPSignIn(passcode: string, targetURL?: string, workflow?: string) {
    const body: CompleteTOTPSigninBody = { token: `${passcode}` };
    if (targetURL) {
        body.targetURL = targetURL;
    }

    if (workflow) {
        body.workflow = workflow;
    }

    return PostWithOptionalResponse<SignInResponse>(CompleteTOTPSignInPath, body);
}
