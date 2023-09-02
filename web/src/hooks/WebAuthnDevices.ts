import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserWebAuthnDevices } from "@services/UserWebAuthnDevices";

export function useUserWebAuthnDevices() {
    return useRemoteCall(getUserWebAuthnDevices, []);
}
