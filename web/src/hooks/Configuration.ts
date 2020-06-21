import { useRemoteCall } from "./RemoteCall";
import { getConfiguration } from "../services/Configuration";

export function useEmbeddedVariable(variableName: string) {
    const value = document.body.getAttribute(`data-${variableName}`);
    if (value === null) {
        throw new Error(`No ${variableName} embedded variable detected`);
    }

    return value;
}

export function useRememberMe() {
    return useEmbeddedVariable("rememberme") === "true";
}

export function useResetPassword() {
    return useEmbeddedVariable("disable-resetpassword") === "true";
}

export function useConfiguration() {
    return useRemoteCall(getConfiguration, []);
}