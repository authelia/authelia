import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import XHR from "i18next-http-backend";
import { initReactI18next } from "react-i18next";

import langEn from "@i18n/locales/en.json";
import langEs from "@i18n/locales/es.json";

const resources = {
    en: langEn,
    es: langEs,
};

const options = {
    order: ["querystring", "navigator"],
    lookupQuerystring: "lng",
};

i18n.use(XHR)
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
        detection: options,
        resources,
        ns: [""],
        defaultNS: "",
        fallbackLng: "en",
        supportedLngs: ["en", "es"],
        interpolation: {
            escapeValue: false,
        },
        debug: false,
    });

export default i18n;
