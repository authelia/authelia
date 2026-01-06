import { useRemoteCall } from "@hooks/RemoteCall";
import {
    deleteUser,
    getAdminConfiguration,
    getAllUserInfo,
    getUser,
    getUserFieldMetadata,
    postNewUser,
    putChangeUser,
} from "@services/UserManagement";

export function useAllUserInfoGET() {
    return useRemoteCall(getAllUserInfo, []);
}

export function useAdminConfigurationGET() {
    return useRemoteCall(getAdminConfiguration, []);
}

export function useUserManagementFieldMetadataGET() {
    return useRemoteCall(getUserFieldMetadata, []);
}

export function useUserGET(username: string) {
    return useRemoteCall(() => getUser(username), [username]);
}

export function useUserPUT() {
    return useRemoteCall(putChangeUser, []);
}

export function useUserPOST() {
    return useRemoteCall(postNewUser, []);
}

export function useUserDELETE() {
    return useRemoteCall(deleteUser, []);
}
