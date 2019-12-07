import { useRemoteCall } from "./RemoteCall";
import { getConfiguration, getExtendedConfiguration } from "../services/Configuration";

export function useConfiguration() {
    return useRemoteCall(getConfiguration, []);
}

export function useExtendedConfiguration() {
    return useRemoteCall(getExtendedConfiguration, []);
}