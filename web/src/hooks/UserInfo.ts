import { useRemoteCall } from "@hooks/RemoteCall";
import { postUserInfo } from "@services/UserInfo";

export function useUserInfoPOST() {
    return useRemoteCall(postUserInfo, []);
}
