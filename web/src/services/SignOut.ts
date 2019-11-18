import { PostWithOptionalResponse } from "./Client";
import { LogoutPath } from "./Api";

export async function signOut() {
    return PostWithOptionalResponse(LogoutPath);
}