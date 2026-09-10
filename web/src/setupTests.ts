// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import "@testing-library/jest-dom";
import i18n from "i18next";
import { initReactI18next } from "react-i18next";

await i18n.use(initReactI18next).init({
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

Object.defineProperty(globalThis, "localStorage", { value: localStorageMock });

document.body.dataset.basepath = "";
document.body.dataset.duoselfenrollment = "true";
document.body.dataset.rememberme = "true";
document.body.dataset.resetpassword = "true";
document.body.dataset.resetpasswordcustomurl = "";
document.body.dataset.totpappapplestore = "https://apps.apple.com/us/app/google-authenticator/id388497605";
document.body.dataset.totpappgoogleplay =
    "https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2";
document.body.dataset.privacypolicyurl = "";
document.body.dataset.privacypolicyaccept = "false";
document.body.dataset.passkeylogin = "true";
document.body.dataset.theme = "light";
