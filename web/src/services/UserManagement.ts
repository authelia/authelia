import { UserInfo } from "@models/UserInfo";
import { AdminConfigPath, AdminManageUserPath, AdminUserInfoPath } from "@services/Api";
import { DeleteWithOptionalResponse, Get, PostWithOptionalResponse, PutWithOptionalResponse } from "@services/Client";
import { UserInfoPayload, toSecondFactorMethod } from "@services/UserInfo";

export async function getAllUserInfo(): Promise<UserInfo[]> {
    const res = await Get<UserInfoPayload[]>(AdminUserInfoPath);
    return res.map((user) => ({ ...user, method: toSecondFactorMethod(user.method) }));
}

export interface AdminConfigBody {
    enabled: boolean;
    admin_group: string;
    allow_admins_to_add_admins: boolean;
}

interface DeleteUserBody {
    username: string;
}

export async function putChangeUser(
    username: string,
    display_name: string,
    password: string | "",
    disabled: boolean | false,
    email: string,
    groups: string[],
) {
    const data = {
        username,
        display_name,
        password,
        disabled,
        email,
        groups,
    };
    return PutWithOptionalResponse(AdminManageUserPath, data);
}

export async function postNewUser(
    username: string,
    display_name: string,
    password: string,
    disabled: boolean,
    email: string,
    groups: string[],
) {
    const data = {
        username,
        display_name,
        password,
        disabled,
        email,
        groups,
    };
    return PostWithOptionalResponse(AdminManageUserPath, data);
}

export async function deleteDeleteUser(username: string) {
    const data: DeleteUserBody = {
        username,
    };
    console.log("delete:", JSON.stringify(data));
    return DeleteWithOptionalResponse(AdminManageUserPath, data);
}

export async function getAdminConfiguration(): Promise<AdminConfigBody> {
    return await Get<AdminConfigBody>(AdminConfigPath);
}
