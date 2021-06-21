import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserPreferences } from "@services/UserPreferences";

export function useUserPreferences() {
    return useRemoteCall(getUserPreferences, []);
}
