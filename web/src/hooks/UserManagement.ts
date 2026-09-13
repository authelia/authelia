import { useRemoteCall } from "@hooks/RemoteCall";
import { getAdminConfiguration, getAllUserInfo, getUser, getUserAttributeMetadata } from "@services/UserManagement";

export function useAllUserInfoGET() {
    return useRemoteCall(getAllUserInfo);
}

export function useAdminConfigurationGET() {
    return useRemoteCall(getAdminConfiguration);
}

export function useUserManagementAttributeMetadataGET() {
    return useRemoteCall(getUserAttributeMetadata);
}

export function useUserGET(username: string) {
    return useRemoteCall(() => getUser(username));
}
