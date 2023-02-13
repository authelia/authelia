import { CompleteTOTPSignInPath, TOTPRegistrationPath } from "@services/Api";
import { DeleteWithOptionalResponse, PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface CompleteTOTPSignInBody {
    token: string;
    targetURL?: string;
    workflow?: string;
    workflowID?: string;
}

export function completeTOTPSignIn(passcode: string, targetURL?: string, workflow?: string, workflowID?: string) {
    const body: CompleteTOTPSignInBody = {
        token: `${passcode}`,
        targetURL: targetURL,
        workflow: workflow,
        workflowID: workflowID,
    };

    return PostWithOptionalResponse<SignInResponse>(CompleteTOTPSignInPath, body);
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
