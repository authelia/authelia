import { LogoutPath } from "./Api";
import { PostWithOptionalResponse } from "./Client";

export type SignOutResponse = { safe_redirect: boolean } | undefined;

export type SignOutBody = {
    redirection_url?: string;
};

export async function signOut(redirectionURL: string | undefined): Promise<SignOutResponse> {
    const body: SignOutBody = {};
    if (redirectionURL) {
        body.redirection_url = redirectionURL;
    }

    return PostWithOptionalResponse<SignOutResponse>(LogoutPath, body);
}
