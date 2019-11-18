import { useRemoteCall } from "./RemoteCall";
import { getAvailable2FAMethods } from "../services/Configuration";

export function useAutheliaConfiguration() {
    return useRemoteCall(getAvailable2FAMethods, []);
}