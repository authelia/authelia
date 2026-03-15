import { UserInfo } from "@models/UserInfo";
import { CreateUserRequest, UserDetailsExtended } from "@models/UserManagement.js";
import { AdminConfigPath, AdminUserAttributeMetadataPath, AdminUserRestPath } from "@services/Api";
import { DeleteWithOptionalResponse, Get, PostWithOptionalResponse, PatchWithOptionalResponse } from "@services/Client";
import { UserInfoPayload, toSecondFactorMethod } from "@services/UserInfo";

export interface AdminConfigBody {
    enabled: boolean;
    admin_group: string;
    allow_admins_to_add_admins: boolean;
}

export type AttributeType = "text" | "email" | "password" | "tel" | "url" | "date";

export interface AttributeMetadata {
    type: AttributeType;
    multiple?: boolean;
}

export interface UserAttributeMetadataBody {
    required_attributes: string[];
    supported_attributes: Record<string, AttributeMetadata>;
}

export async function getAdminConfiguration(): Promise<AdminConfigBody> {
    return await Get<AdminConfigBody>(AdminConfigPath);
}

export async function getUserAttributeMetadata(): Promise<UserAttributeMetadataBody> {
    return await Get<UserAttributeMetadataBody>(AdminUserAttributeMetadataPath);
}

export async function getAllUserInfo(): Promise<UserInfo[]> {
    const res = await Get<UserInfoPayload[]>(AdminUserRestPath);
    return res.map((user) => ({ ...user, method: toSecondFactorMethod(user.method) }));
}

export async function getUser(username: string): Promise<UserDetailsExtended> {
    return await Get<UserDetailsExtended>(`${AdminUserRestPath}/${username}`);
}

export async function patchChangeUser(username: string, userData: Partial<UserDetailsExtended>, updateMask: string[]) {
    const data = { ...userData };
    if (!data.password) {
        delete data.password;
    }

    // Build the update_mask query parameter
    const updateMaskParam = updateMask.join(',');

    return PatchWithOptionalResponse(`${AdminUserRestPath}/${username}?update_mask=${updateMaskParam}`, data);
}

// postNewUser uses the rest api to create a new user.
export async function postNewUser(userData: UserDetailsExtended) {
    return PostWithOptionalResponse(AdminUserRestPath, userData);
}

// deleteUser uses the rest api to delete an existing user.
export async function deleteUser(username: string) {
    return DeleteWithOptionalResponse(`${AdminUserRestPath}/${username}`);
}
