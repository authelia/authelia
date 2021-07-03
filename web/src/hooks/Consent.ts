import { useRemoteCall } from "@hooks/RemoteCall";
import { getRequestedScopes } from "@services/Consent";

export function useRequestedScopes() {
    return useRemoteCall(getRequestedScopes, []);
}
