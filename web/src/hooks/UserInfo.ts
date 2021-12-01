import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserInfo } from "@services/UserInfo";

export function useUserInfo() {
    return useRemoteCall(getUserInfo, []);
}
