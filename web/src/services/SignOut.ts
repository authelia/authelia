import { LogoutPath } from "./Api";
import { PostWithOptionalResponse } from "./Client";

export type SignOutResponse = { safeTargetURL: boolean } | undefined;

export type SignOutBody = {
    targetURL?: string;
};

export async function signOut(targetURL: string | undefined): Promise<SignOutResponse> {
    const body: SignOutBody = {};
    if (targetURL) {
        body.targetURL = targetURL;
    }

    return PostWithOptionalResponse<SignOutResponse>(LogoutPath, body);
}
