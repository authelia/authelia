import { CompleteTOTPSignInPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface CompleteTOTPSigninBody {
    token: string;
    targetURL?: string;
}

export function completeTOTPSignIn(passcode: string, targetURL: string | undefined) {
    const body: CompleteTOTPSigninBody = { token: `${passcode}` };
    if (targetURL) {
        body.targetURL = targetURL;
    }
    return PostWithOptionalResponse<SignInResponse>(CompleteTOTPSignInPath, body);
}
