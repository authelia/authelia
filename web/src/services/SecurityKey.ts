import u2fApi from "u2f-api";

import { InitiateU2FSignInPath, CompleteU2FSignInPath } from "./Api";
import { Post, PostWithOptionalResponse } from "./Client";
import { SignInResponse } from "./SignIn";

interface InitiateU2FSigninResponse {
    appId: string;
    challenge: string;
    registeredKeys: {
        appId: string;
        keyHandle: string;
        version: string;
    }[];
}

export async function initiateU2FSignin() {
    return Post<InitiateU2FSigninResponse>(InitiateU2FSignInPath);
}

interface CompleteU2FSigninBody {
    signResponse: u2fApi.SignResponse;
    targetURL?: string;
}

export function completeU2FSignin(signResponse: u2fApi.SignResponse, targetURL: string | undefined) {
    const body: CompleteU2FSigninBody = { signResponse };
    if (targetURL) {
        body.targetURL = targetURL;
    }
    return PostWithOptionalResponse<SignInResponse>(CompleteU2FSignInPath, body);
}
