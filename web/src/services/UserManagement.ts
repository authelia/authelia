import { UserDetailsExtended } from "@models/UserManagement.js";
import {
    AdminChangePasswordRestSubPath,
    AdminConfigPath,
    AdminSendResetPasswordEmailRestSubPath,
    AdminUserAttributeMetadataPath,
    AdminUserRestPath,
} from "@services/Api";
import { DeleteWithOptionalResponse, Get, PatchWithOptionalResponse, PostWithOptionalResponse } from "@services/Client";

export interface AdminConfigBody {
    enabled: boolean;
    admin_group: string;
    allow_admins_to_add_admins: boolean;
    group_management_enabled: boolean;
}

export type AttributeType = "checkbox" | "date" | "email" | "groups" | "number" | "password" | "tel" | "text" | "url";

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

export async function getAllUserInfo(): Promise<UserDetailsExtended[]> {
    return await Get<UserDetailsExtended[]>(AdminUserRestPath);
}

export async function getUser(username: string): Promise<UserDetailsExtended> {
    return await Get<UserDetailsExtended>(`${AdminUserRestPath}/${username}`);
}

export async function patchChangeUser(username: string, userData: Partial<UserDetailsExtended>, updateMask: string[]) {
    const data = { ...userData };
    if (!data.password) {
        delete data.password;
    }

    const updateMaskParam = updateMask.join(",");

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

export async function postChangePasswordForUser(username: string, password: string) {
    return PostWithOptionalResponse(`${AdminUserRestPath}/${username}${AdminChangePasswordRestSubPath}`, { password });
}

export async function postSendResetPasswordEmailForUser(username: string) {
    return PostWithOptionalResponse(`${AdminUserRestPath}/${username}${AdminSendResetPasswordEmailRestSubPath}`);
}
