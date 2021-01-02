import { CompleteTOTPSignInPath } from "./Api";
import { PostWithOptionalResponse } from "./Client";
import { SignInResponse } from "./SignIn";

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
