import i18n from "i18next";
import BrowserLanguageDetector from "i18next-browser-languagedetector";
import HTTPBackend from "i18next-http-backend";
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

i18n.use(HTTPBackend)
    .use(BrowserLanguageDetector)
    .use(initReactI18next)
    .init({
        detection: options,
        resources,
        ns: [""],
        defaultNS: "",
        fallbackLng: {
            default: ["en"],
        },
        supportedLngs: [
            "en",
            "en-AU",
            "en-BZ",
            "en-CA",
            "en-IE",
            "en-JM",
            "en-NZ",
            "en-ZA",
            "en-TT",
            "en-GB",
            "en-US",
            "es",
            "es-AR",
            "es-BO",
            "es-CL",
            "es-CO",
            "es-CR",
            "es-DO",
            "es-EC",
            "es-SV",
            "es-GT",
            "es-HN",
            "es-MX",
            "es-NI",
            "es-PA",
            "es-PY",
            "es-PE",
            "es-PR",
            "es-UY",
            "es-VE",
        ],
        interpolation: {
            escapeValue: false,
        },
        debug: true,
    });

export default i18n;
