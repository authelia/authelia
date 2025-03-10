import { CustomDetector, DetectorOptions } from "i18next-browser-languagedetector";

import { LocalStorageLanguagePreference } from "@constants/LocalStorage";
import { getLocalStorage, localStorageAvailable } from "@services/LocalStorage";

const LocalStorageCustomDetector: CustomDetector = {
    name: "localStorageCustom",

    lookup(options: DetectorOptions): string | undefined {
        let found;

        if (options.lookupLocalStorage && localStorageAvailable()) {
            const lng = getLocalStorage(LocalStorageLanguagePreference);

            if (lng && lng !== "") {
                found = lng;
            }
        }

        return found;
    },
};

export default LocalStorageCustomDetector;
