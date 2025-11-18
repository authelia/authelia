let hasLocalStorageSupport: boolean | null = null;
const testKey = "authelia.test";
const testValue = "foo";

export function localStorageAvailable() {
    if (hasLocalStorageSupport !== null) return hasLocalStorageSupport;

    hasLocalStorageSupport = false;

    if (typeof globalThis !== "undefined" && globalThis.localStorage !== null) {
        hasLocalStorageSupport = true;

        try {
            globalThis.localStorage.setItem(testKey, testValue);
            globalThis.localStorage.removeItem(testKey);
        } catch {
            hasLocalStorageSupport = false;
        }
    }

    return hasLocalStorageSupport;
}

export function getLocalStorage(key: string) {
    if (!localStorageAvailable()) return null;

    return globalThis.localStorage.getItem(key);
}

export function setLocalStorage(key: string, value: string) {
    if (!localStorageAvailable()) return false;

    globalThis.localStorage.setItem(key, value);

    return true;
}
