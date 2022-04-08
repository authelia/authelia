import { useRemoteCall } from "@hooks/RemoteCall";
import { getConsentResponse } from "@services/Consent";

export function useConsentResponse() {
    return useRemoteCall(getConsentResponse, []);
}
