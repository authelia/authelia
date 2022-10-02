import { LogoutPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";

export type SignOutResponse = { safeTargetURL: boolean } | undefined;

export type SignOutBody = {
    targetURL?: string;
};

export async function signOut(targetURL: string | null): Promise<SignOutResponse> {
    const body: SignOutBody = {
        targetURL: targetURL === null ? undefined : targetURL,
    };

    return PostWithOptionalResponse<SignOutResponse>(LogoutPath, body);
}
