import { CustomDetector, DetectorOptions } from "i18next-browser-languagedetector";

import { localStorageAvailable } from "@utils/localStorage";

const LocalStorageCustomDetector: CustomDetector = {
    name: "localStorageCustom",

    lookup(options: DetectorOptions): string | undefined {
        let found;

        if (options.lookupLocalStorage && localStorageAvailable()) {
            const lng = window.localStorage.getItem(options.lookupLocalStorage);
            if (lng && lng !== "") {
                found = lng;
            }
        }

        return found;
    },
};

export default LocalStorageCustomDetector;
