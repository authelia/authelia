import { CompletePushNotificationSignInPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

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
