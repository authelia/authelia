import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserWebauthnDevices } from "@services/UserWebauthnDevices";

export function useUserWebauthnDevices() {
    return useRemoteCall(getUserWebauthnDevices, []);
}
