import { CompleteTOTPSignInPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface CompleteU2FSigninBody {
    token: string;
    targetURL?: string;
}

export function completeTOTPSignIn(passcode: string, targetURL: string | undefined) {
    const body: CompleteU2FSigninBody = { token: `${passcode}` };
    if (targetURL) {
        body.targetURL = targetURL;
    }
    return PostWithOptionalResponse<SignInResponse>(CompleteTOTPSignInPath, body);
}
