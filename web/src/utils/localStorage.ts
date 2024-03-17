let hasLocalStorageSupport: null | boolean = null;
const testKey = "authelia.test";
const testValue = "foo";

export function localStorageAvailable() {
    if (hasLocalStorageSupport !== null) return hasLocalStorageSupport;

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

export function localStoreSet(key: string, value: string) {
    if (localStorageAvailable()) {
        window.localStorage.setItem(key, value);
    } else {
        console.error("local storage not supported");
    }
}
