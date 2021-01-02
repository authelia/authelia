import { getConfiguration } from "../services/Configuration";
import { useRemoteCall } from "./RemoteCall";

export function useConfiguration() {
    return useRemoteCall(getConfiguration, []);
}
