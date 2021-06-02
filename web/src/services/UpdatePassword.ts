import { UpdatePasswordPath } from "./Api";
import { PostWithOptionalResponse } from "./Client";

export async function updatePassword(oldPassword: string, newPassword: string) {
    return PostWithOptionalResponse(UpdatePasswordPath, { old_password: oldPassword, password: newPassword });
}
