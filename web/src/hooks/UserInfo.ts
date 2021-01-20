import { getUserPreferences } from "../services/UserPreferences";
import { useRemoteCall } from "./RemoteCall";

export function useUserPreferences() {
    return useRemoteCall(getUserPreferences, []);
}
