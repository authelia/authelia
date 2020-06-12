import { useRemoteCall } from "./RemoteCall";
import { getRequestedScopes } from "../services/Consent";

export function useRequestedScopes() {
    return useRemoteCall(getRequestedScopes, []);
}