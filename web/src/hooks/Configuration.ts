import { useRemoteCall } from "./RemoteCall";
import { getConfiguration } from "../services/Configuration";

export function useRememberMe() {
    const rememberMe = (document.body.getAttribute("data-rememberme") === 'true');
    if (rememberMe === null) {
        throw new Error("No remember me setting detected");
    }

    return rememberMe;
}

export function useResetPassword() {
    const resetPassword = (document.body.getAttribute("data-disable-resetpassword") === 'true');
    if (resetPassword === null) {
        throw new Error("No reset password setting detected");
    }

    return resetPassword;
}

export function useConfiguration() {
    return useRemoteCall(getConfiguration, []);
}