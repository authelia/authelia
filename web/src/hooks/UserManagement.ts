import { useRemoteCall } from "@hooks/RemoteCall";
import { getAdminConfiguration, getAllUserInfo } from "@services/UserManagement";

export function useAllUserInfoGET() {
    return useRemoteCall(getAllUserInfo, []);
}

export function useAdminConfigurationGET() {
    return useRemoteCall(getAdminConfiguration, []);
}
