import { DetectorOptions } from "i18next-browser-languagedetector";

let hasLocalStorageSupport: null | boolean = null;
const testKey = 'authelia.test';

const localStorageAvailable = () => {
    if (hasLocalStorageSupport !== null) return hasLocalStorageSupport;

    if (typeof window !== 'undefined' && window.localStorage !== null) {
        hasLocalStorageSupport = true;

        try {
            window.localStorage.setItem(testKey, 'foo');
            window.localStorage.removeItem(testKey);
        } catch (e) {
            hasLocalStorageSupport = false;
        }
    }

    return hasLocalStorageSupport;
};

export default {
    name: 'localStorageCustom',

    lookup(options: DetectorOptions): string | undefined {
        let found;

        if (options.lookupLocalStorage && localStorageAvailable()) {
            const lng = window.localStorage.getItem(options.lookupLocalStorage);
            if (lng && lng !== '') {
                found = lng;
            }
        }

        return found;
    },
};
