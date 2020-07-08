import { useRemoteCall } from "./RemoteCall";
import { getConfiguration } from "../services/Configuration";

export function useConfiguration() {
    return useRemoteCall(getConfiguration, []);
}