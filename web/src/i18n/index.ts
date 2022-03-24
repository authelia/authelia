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
        fallbackLng: {
            default: ["en"],
            "en-AU": ["en"],
            "en-BZ": ["en"],
            "en-CA": ["en"],
            "en-IE": ["en"],
            "en-JM": ["en"],
            "en-NZ": ["en"],
            "en-ZA": ["en"],
            "en-TT": ["en"],
            "en-GB": ["en"],
            "en-US": ["en"],
            "es-AR": ["es"],
            "es-BO": ["es"],
            "es-CL": ["es"],
            "es-CO": ["es"],
            "es-CR": ["es"],
            "es-DO": ["es"],
            "es-EC": ["es"],
            "es-SV": ["es"],
            "es-GT": ["es"],
            "es-HN": ["es"],
            "es-MX": ["es"],
            "es-NI": ["es"],
            "es-PA": ["es"],
            "es-PY": ["es"],
            "es-PE": ["es"],
            "es-PR": ["es"],
            "es-UY": ["es"],
            "es-VE": ["es"],
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
        debug: false,
    });

export default i18n;
