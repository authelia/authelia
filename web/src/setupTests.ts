import matchers, { TestingLibraryMatchers } from "@testing-library/jest-dom/matchers";

declare global {
    namespace Vi {
        interface JestAssertion<T = any> extends jest.Matchers<void, T>, TestingLibraryMatchers<T, void> {}
    }
}

expect.extend(matchers);

const localStorageMock = (function () {
    let store = {};

    return {
        getItem(key) {
            return store[key];
        },

        setItem(key, value) {
            store[key] = value;
        },

        clear() {
            store = {};
        },

        removeItem(key) {
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
