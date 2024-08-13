import { ChangePasswordPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";

interface PostPasswordChange {
    username: string;
    oldPassword: string;
    newPassword: string;
}

export async function postPasswordChange(username: string, oldPassword: string, newPassword: string) {
    const data: PostPasswordChange = {
        username,
        oldPassword,
        newPassword,
    };
    return PostWithOptionalResponse(ChangePasswordPath, data);
}
