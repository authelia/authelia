import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserInfo, postUserInfo } from "@services/UserInfo";

export function useUserInfo() {
    return useRemoteCall(getUserInfo, []);
}

export function useUserInfoPOST() {
    return useRemoteCall(postUserInfo, []);
}
