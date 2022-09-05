import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import Backend from "i18next-http-backend";
import { initReactI18next } from "react-i18next";

import { getBasePath } from "@utils/BasePath";

import LocalStorageCustomDetector from "./localStorageCustom";

const basePath = getBasePath();

const detector = new LanguageDetector();

detector.addDetector(LocalStorageCustomDetector);

i18n.use(Backend)
    .use(detector)
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
            "nl-NL": ["en"],
            "pt-PT": ["en"],
            ru: ["en"],
            sv: ["en"],
            "sv-SE": ["sv", "en"],
            "zh-CN": ["en"],
            "zh-TW": ["en"],
        },
        supportedLngs: ["en", "de", "es", "fr", "nl-NL", "pt-PT", "ru", "sv", "sv-SE", "zh-CN", "zh-TW"],
        lowerCaseLng: false,
        nonExplicitSupportedLngs: true,
        interpolation: {
            escapeValue: false,
        },
        debug: false,
    });

export default i18n;
