import { getRequestedScopes } from "../services/Consent";
import { useRemoteCall } from "./RemoteCall";


export function useRequestedScopes() {
    return useRemoteCall(getRequestedScopes, []);
}
