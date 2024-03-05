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

export function getLocalStorage(key: string) {
    if (!localStorageAvailable()) return null;

    return window.localStorage.getItem(key);
}
