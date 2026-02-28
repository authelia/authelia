import { FirstFactorSpnego } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { SignInResponse } from "@services/SignIn";

interface SpnegoPostFirstFactorBody {
    keepMeLoggedIn: boolean;
    targetURL?: string;
    requestMethod?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
    userCode?: string;
}

export async function spnegoPostFirstFactor(
    rememberMe: boolean,
    targetURL?: string,
    requestMethod?: string,
    flowID?: string,
    flow?: string,
    subflow?: string,
    userCode?: string,
) {
    const data: SpnegoPostFirstFactorBody = {
        flow,
        flowID,
        keepMeLoggedIn: rememberMe,
        requestMethod,
        subflow,
        targetURL,
        userCode,
    };

    const res = await PostWithOptionalResponse<SignInResponse>(FirstFactorSpnego, data);
    return res ?? ({} as SignInResponse);
}
