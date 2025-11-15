import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserWebAuthnCredentials } from "@services/UserWebAuthnCredentials";

export function useUserWebAuthnCredentials() {
    return useRemoteCall(getUserWebAuthnCredentials);
}
