import "@testing-library/jest-dom";
import i18n from "i18next";
import { initReactI18next } from "react-i18next";

i18n.use(initReactI18next).init({
    resources: { en: { testNS: {} } },
});

interface LocalStorageMock {
    [key: string]: any;
}

const localStorageMock: LocalStorageMock = (function () {
    let store: LocalStorageMock = {};

    return {
        getItem(key: string | number) {
            return store[key];
        },

        setItem(key: string | number, value: any) {
            store[key] = value;
        },

        clear() {
            store = {};
        },

        removeItem(key: string | number) {
            delete store[key];
        },

        getAll() {
            return store;
        },
    };
})();

Object.defineProperty(window, "localStorage", { value: localStorageMock });

document.body.setAttribute("data-basepath", "");
document.body.setAttribute("data-duoselfenrollment", "true");
document.body.setAttribute("data-rememberme", "true");
document.body.setAttribute("data-resetpassword", "true");
document.body.setAttribute("data-resetpasswordcustomurl", "");
document.body.setAttribute("data-privacypolicyurl", "");
document.body.setAttribute("data-privacypolicyaccept", "false");
document.body.setAttribute("data-theme", "light");
