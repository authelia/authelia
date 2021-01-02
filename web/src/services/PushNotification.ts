import { CompletePushNotificationSignInPath } from "./Api";
import { PostWithOptionalResponse } from "./Client";
import { SignInResponse } from "./SignIn";

interface CompleteU2FSigninBody {
    targetURL?: string;
}

export function completePushNotificationSignIn(targetURL: string | undefined) {
    const body: CompleteU2FSigninBody = {};
    if (targetURL) {
        body.targetURL = targetURL;
    }
    return PostWithOptionalResponse<SignInResponse>(CompletePushNotificationSignInPath, body);
}
