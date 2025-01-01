import { ChangePasswordPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";

interface PostPasswordChange {
    username: string;
    old_password: string;
    new_password: string;
}

export async function postPasswordChange(username: string, old_password: string, new_password: string) {
    const data: PostPasswordChange = {
        username,
        old_password,
        new_password,
    };
    return PostWithOptionalResponse(ChangePasswordPath, data);
}
