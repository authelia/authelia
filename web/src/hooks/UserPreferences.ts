import { getUserPreferences } from "../services/UserPreferences";
import { useRemoteCall } from "../hooks/RemoteCall";

export function useUserPreferences() {
    return useRemoteCall(getUserPreferences, []);
}