import { LogoutPath } from "./Api";
import { PostWithOptionalResponse } from "./Client";

export async function signOut() {
    return PostWithOptionalResponse(LogoutPath);
}
