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
        clear() {
            store = {};
        },

        getAll() {
            return store;
        },

        getItem(key: number | string) {
            return store[key];
        },

        removeItem(key: number | string) {
            delete store[key];
        },

        setItem(key: number | string, value: any) {
            store[key] = value;
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
document.body.setAttribute("data-passkeylogin", "true");
document.body.setAttribute("data-theme", "light");
