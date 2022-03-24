import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import Backend from "i18next-http-backend";
import { initReactI18next } from "react-i18next";

import { getConfigurationLanguages } from "@services/Configuration";

const config = getConfigurationLanguages();

i18n.use(Backend)
    .use(LanguageDetector)
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
        supportedLngs: config.supported_languages,
        nonExplicitSupportedLngs: true,
        interpolation: {
            escapeValue: false,
        },
        debug: true,
    });

export default i18n;
