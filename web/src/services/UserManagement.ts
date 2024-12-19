import { UserInfo } from "@models/UserInfo";
import { AdminConfigPath, AdminManageUserPath, AdminUserInfoPath } from "@services/Api";
import { DeleteWithOptionalResponse, Get, PostWithOptionalResponse, PutWithOptionalResponse } from "@services/Client";
import { UserInfoPayload, toSecondFactorMethod } from "@services/UserInfo";

export async function getAllUserInfo(): Promise<UserInfo[]> {
    const res = await Get<UserInfoPayload[]>(AdminUserInfoPath);
    return res.map((user) => ({ ...user, method: toSecondFactorMethod(user.method) }));
}

interface UserChangeBody {
    username: string;
    display_name: string;
    email: string;
    groups: string[];
}

export interface AdminConfigBody {
    enabled: boolean;
    admin_group: string;
    allow_admins_to_add_admins: boolean;
}

interface DeleteUserBody {
    username: string;
}

export async function postChangeUser(username: string, display_name: string, email: string, groups: string[]) {
    const data: UserChangeBody = {
        username,
        display_name,
        email,
        groups,
    };
    return PostWithOptionalResponse(AdminManageUserPath, data);
}

export async function putNewUser(
    username: string,
    display_name: string,
    password: string,
    email: string,
    groups: string[],
) {
    const data = {
        username,
        display_name,
        password,
        email,
        groups,
    };
    return PutWithOptionalResponse(AdminManageUserPath, data);
}

export async function deleteDeleteUser(username: string) {
    const data: DeleteUserBody = {
        username,
    };
    return DeleteWithOptionalResponse(AdminManageUserPath, data);
}

export async function getAdminConfiguration(): Promise<AdminConfigBody> {
    return await Get<AdminConfigBody>(AdminConfigPath);
}
