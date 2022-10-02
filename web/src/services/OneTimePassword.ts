import { CompleteTOTPSignInPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface CompleteTOTPSigninBody {
    token: string;
    targetURL?: string;
    workflow?: string;
}

export function completeTOTPSignIn(passcode: string, targetURL: string | null, workflow: string | null) {
    const body: CompleteTOTPSigninBody = {
        token: `${passcode}`,
        targetURL: targetURL === null ? undefined : targetURL,
        workflow: workflow === null ? undefined : workflow,
    };

    return PostWithOptionalResponse<SignInResponse>(CompleteTOTPSignInPath, body);
}
