import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import Backend from "i18next-http-backend";
import { initReactI18next } from "react-i18next";

import { getBasePath } from "@utils/BasePath";

const basePath = getBasePath();

i18n.use(Backend)
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
        detection: {
            order: ["querystring", "navigator"],
            lookupQuerystring: "lng",
        },
        backend: {
            loadPath: basePath + "/locales/{{lng}}/{{ns}}.json",
        },
        //  portal: namespace for frontend translations
        //  backend: namespace for api translations
        ns: ["portal", "backend"],
        defaultNS: "portal",
        fallbackLng: {
            default: ["en"],
        },
        load: "all",
        supportedLngs: ["en", "es", "de"],
        lowerCaseLng: false,
        nonExplicitSupportedLngs: true,
        interpolation: {
            escapeValue: false,
        },
        debug: false,
    });

export default i18n;
