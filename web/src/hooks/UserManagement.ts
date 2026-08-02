import { useRemoteCall } from "@hooks/RemoteCall";
import {
    deleteUser,
    getAdminConfiguration,
    getAllUserInfo,
    getUser,
    getUserAttributeMetadata,
    patchChangeUser,
    postChangePasswordForUser,
    postNewUser,
    postSendResetPasswordEmailForUser,
} from "@services/UserManagement";

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

export function useUserPATCH() {
    return useRemoteCall(patchChangeUser);
}

export function useUserPOST() {
    return useRemoteCall(postNewUser);
}

export function useUserDELETE() {
    return useRemoteCall(deleteUser);
}

export function useChangePasswordForUserPOST(username: string, password: string) {
    return useRemoteCall(() => postChangePasswordForUser(username, password));
}

export function useSendResetPasswordEmailForUserPOST(username: string) {
    return useRemoteCall(() => postSendResetPasswordEmailForUser(username));
}
