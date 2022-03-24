import i18n from "i18next";
import BrowserLanguageDetector from "i18next-browser-languagedetector";
import HTTPBackend from "i18next-http-backend";
import { initReactI18next } from "react-i18next";

i18n.use(HTTPBackend)
    .use(BrowserLanguageDetector)
    .use(initReactI18next)
    .init({
        detection: {
            order: ["querystring", "navigator"],
            lookupQuerystring: "locale",
        },
        backend: {
            loadPath: "/locale.json?lng={{lng}}&ns={{ns}}",
        },
        ns: ["portal"],
        defaultNS: "portal",
        fallbackLng: {
            default: ["en"],
        },
        load: "all",
        supportedLngs: ["en", "es"],
        nonExplicitSupportedLngs: true,
        interpolation: {
            escapeValue: false,
        },
        debug: true,
    });

export default i18n;
