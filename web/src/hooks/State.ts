import { useRemoteCall } from "@hooks/RemoteCall";
import { getState } from "@services/State";

export function useAutheliaState() {
    return useRemoteCall(getState);
}
