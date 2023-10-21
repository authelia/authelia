import { CustomDetector, DetectorOptions } from "i18next-browser-languagedetector";

import { getLocalStorage } from "@services/LocalStorage";

const LocalStorageCustomDetector: CustomDetector = {
    name: "localStorageCustom",

    lookup(options: DetectorOptions): string | undefined {
        let found;

        if (options.lookupLocalStorage) {
            const lng = getLocalStorage(options.lookupLocalStorage);

            if (lng && lng !== "") {
                found = lng;
            }
        }

        return found;
    },
};

export default LocalStorageCustomDetector;
