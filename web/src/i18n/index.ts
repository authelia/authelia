import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import Backend from "i18next-http-backend";
import { initReactI18next } from "react-i18next";

import LocalStorageCustomDetector from "@i18n/detectors/localStorageCustom";
import { getBasePath } from "@utils/BasePath";

const basePath = getBasePath();

const CustomLanguageDetector = new LanguageDetector().addDetector(LocalStorageCustomDetector);

i18n.use(Backend)
    .use(CustomLanguageDetector)
    .use(initReactI18next)
    .init({
        detection: {
            order: ["querystring", "localStorageCustom", "navigator"],
            lookupQuerystring: "lng",
            lookupLocalStorage: "lng",
        },
        backend: {
            loadPath: basePath + "/locales/{{lng}}/{{ns}}.json",
        },
        ns: ["portal"],
        defaultNS: "portal",
        load: "all",
        fallbackLng: {
            default: ["en"],
            de: ["en"],
            es: ["en"],
            fr: ["en"],
            nl: ["en"],
            pt: ["en"],
            ru: ["en"],
            sv: ["en"],
            "sv-SE": ["sv", "en"],
            zh: ["en"],
            "zh-CN": ["zh", "en"],
            "zh-TW": ["zh", "en"],
        },
        supportedLngs: ["en", "de", "es", "fr", "nl", "pt", "ru", "sv", "sv-SE", "zh", "zh-CN", "zh-TW"],
        lowerCaseLng: false,
        nonExplicitSupportedLngs: true,
        interpolation: {
            escapeValue: false,
        },
        debug: false,
    });

export default i18n;
