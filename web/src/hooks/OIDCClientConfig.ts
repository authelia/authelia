import { useRemoteCall } from "@hooks/RemoteCall";
import { getOpenIDConnectClients } from "@services/OIDCClientConfig";

export function useOpenIDConnectClients() {
    return useRemoteCall(getOpenIDConnectClients, []);
}
