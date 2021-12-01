import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserInfoTOTPConfiguration } from "@services/UserInfoTOTPConfiguration";

export function useUserInfoTOTPConfiguration() {
    return useRemoteCall(getUserInfoTOTPConfiguration, []);
}
