import { UserInfo } from "@models/UserInfo";
import { CreateUserRequest, UserDetailsExtended } from "@models/UserManagement.js";
import { AdminConfigPath, AdminUserFieldMetadataPath, AdminUserRestPath } from "@services/Api";
import { DeleteWithOptionalResponse, Get, PostWithOptionalResponse, PutWithOptionalResponse } from "@services/Client";
import { UserInfoPayload, toSecondFactorMethod } from "@services/UserInfo";

export interface AdminConfigBody {
    enabled: boolean;
    admin_group: string;
    allow_admins_to_add_admins: boolean;
}

export type FieldType = "array" | "email" | "password" | "string" | "url";

export interface FieldMetadata {
    display_name: string;
    description: string;
    type: FieldType;
    maxLength?: number;
    pattern?: string;
}

export interface UserFieldMetadataBody {
    required_fields: (keyof CreateUserRequest)[];
    supported_fields: (keyof UserDetailsExtended)[];
    field_metadata: Record<keyof UserDetailsExtended, FieldMetadata>;
}

export async function getAdminConfiguration(): Promise<AdminConfigBody> {
    return await Get<AdminConfigBody>(AdminConfigPath);
}

export async function getUserFieldMetadata(): Promise<UserFieldMetadataBody> {
    return await Get<UserFieldMetadataBody>(AdminUserFieldMetadataPath);
}

export async function getAllUserInfo(): Promise<UserInfo[]> {
    const res = await Get<UserInfoPayload[]>(AdminUserRestPath);
    return res.map((user) => ({ ...user, method: toSecondFactorMethod(user.method) }));
}

export async function getUser(username: string): Promise<UserDetailsExtended> {
    return await Get<UserDetailsExtended>(`${AdminUserRestPath}/${username}`);
}

export async function putChangeUser(username: string, userData: Partial<UserDetailsExtended>) {
    const data = { ...userData };
    if (!data.password) {
        delete data.password;
    }

    return PutWithOptionalResponse(`${AdminUserRestPath}/${username}`, data);
}

// postNewUser uses the rest api to create a new user.
export async function postNewUser(userData: UserDetailsExtended) {
    return PostWithOptionalResponse(AdminUserRestPath, userData);
}

// deleteUser uses the rest api to delete an existing user.
export async function deleteUser(username: string) {
    return DeleteWithOptionalResponse(`${AdminUserRestPath}/${username}`);
}
