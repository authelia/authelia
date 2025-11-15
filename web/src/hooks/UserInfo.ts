import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserInfo, postUserInfo } from "@services/UserInfo";

export function useUserInfoPOST() {
    return useRemoteCall(postUserInfo);
}

export function useUserInfoGET() {
    return useRemoteCall(getUserInfo);
}
