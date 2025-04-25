import { CompleteTOTPSignInPath, TOTPRegistrationPath } from "@services/Api";
import {
    DeleteWithOptionalResponse,
    PostWithOptionalResponse,
    PostWithOptionalResponseRateLimited,
} from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface CompleteTOTPSignInBody {
    token: string;
    targetURL?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
}

export function completeTOTPSignIn(
    passcode: string,
    targetURL?: string,
    flowID?: string,
    flow?: string,
    subflow?: string,
) {
    const body: CompleteTOTPSignInBody = {
        token: `${passcode}`,
        targetURL,
        flowID,
        flow,
        subflow,
    };

    return PostWithOptionalResponseRateLimited<SignInResponse>(CompleteTOTPSignInPath, body);
}

export function completeTOTPRegister(passcode: string) {
    const body: CompleteTOTPSignInBody = {
        token: `${passcode}`,
    };

    return PostWithOptionalResponse(TOTPRegistrationPath, body);
}

export function stopTOTPRegister() {
    return DeleteWithOptionalResponse(TOTPRegistrationPath);
}
