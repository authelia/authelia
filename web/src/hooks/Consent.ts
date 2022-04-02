import { useRemoteCall } from "@hooks/RemoteCall";
import { getConsentRequest } from "@services/Consent";

export function useConsentRequest() {
    return useRemoteCall(getConsentRequest, []);
}
