import i18n, { FallbackLngObjList } from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import Backend from "i18next-http-backend";
import { initReactI18next } from "react-i18next";

import { LocalStorageLanguageCurrent } from "@constants/LocalStorage";
import LocalStorageCustomDetector from "@i18n/detectors/localStorageCustom";
import { getLocaleInformation } from "@services/LocaleInformation";
import { getBasePath } from "@utils/BasePath";

const basePath = getBasePath();

const CustomLanguageDetector = new LanguageDetector();

CustomLanguageDetector.addDetector(LocalStorageCustomDetector);

i18n.use(Backend)
    .use(CustomLanguageDetector)
    .use(initReactI18next)
    .init({
        detection: {
            order: ["querystring", "localStorageCustom", "navigator"],
            lookupQuerystring: "lng",
            lookupLocalStorage: LocalStorageLanguageCurrent,
        },
        backend: {
            loadPath: basePath + "/locales/{{lng}}/{{ns}}.json",
        },
        load: "all",
        ns: ["portal", "settings"],
        defaultNS: "portal",
        fallbackLng: {
            default: ["en"],
        },
        lowerCaseLng: false,
        nonExplicitSupportedLngs: true,
        interpolation: {
            escapeValue: false,
        },
        debug: false,
    });

export default i18n;

getLocaleInformation()
    .then((response) => {
        const supportedLngs = response.languages.map((l) => l.locale);
        var fallbackLng: FallbackLngObjList = {
            default: ["en"],
        };
        response.languages.forEach((l) => {
            fallbackLng[l.locale] = l.fallbacks;
        });
        i18n.options.supportedLngs = supportedLngs;
        i18n.options.fallbackLng = fallbackLng;
    })
    .catch((err) => {
        console.error(err);
    });
