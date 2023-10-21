import { LocalStorageSecondFactorMethod } from "@constants/LocalStorage";
import { SecondFactorMethod } from "@models/Methods";
import { Method2FA, isMethod2FA, toMethod2FA, toSecondFactorMethod } from "@services/UserInfo";

let hasLocalStorageSupport: null | boolean = null;
const testKey = "authelia.test";
const testValue = "foo";

export function localStorageAvailable() {
    if (hasLocalStorageSupport !== null) return hasLocalStorageSupport;

    hasLocalStorageSupport = false;

    if (typeof window !== "undefined" && window.localStorage !== null) {
        hasLocalStorageSupport = true;

        try {
            window.localStorage.setItem(testKey, testValue);
            window.localStorage.removeItem(testKey);
        } catch (e) {
            hasLocalStorageSupport = false;
        }
    }

    return hasLocalStorageSupport;
}

export function removeLocalStorage(key: string) {
    if (!localStorageAvailable()) return false;

    window.localStorage.removeItem(key);

    return true;
}

export function getLocalStorage(key: string) {
    if (!localStorageAvailable()) return null;

    return window.localStorage.getItem(key);
}

export function setLocalStorage(key: string, value: string) {
    if (!localStorageAvailable()) return false;

    window.localStorage.setItem(key, value);

    return true;
}

export function setLocalStorageSecondFactorMethod(value: SecondFactorMethod): boolean {
    return setLocalStorage(LocalStorageSecondFactorMethod, toMethod2FA(value));
}

export function getLocalStorageSecondFactorMethod(global: SecondFactorMethod): SecondFactorMethod {
    const method = getLocalStorage(LocalStorageSecondFactorMethod);

    if (method === null) return global;

    if (!isMethod2FA(method)) {
        return global;
    }

    const local: Method2FA = method as "webauthn" | "totp" | "mobile_push";

    return toSecondFactorMethod(local);
}
