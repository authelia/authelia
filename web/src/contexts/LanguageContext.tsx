import React, { createContext, useCallback, useContext, useEffect, useState } from "react";

import { i18n } from "i18next";

import { LocalStorageLanguagePreference } from "@constants/LocalStorage";
import { setLocalStorage } from "@services/LocalStorage";

export const LanguageContext = createContext<null | ValueProps>(null);

export interface Props {
    i18n: i18n;
    children: React.ReactNode;
}

export interface ValueProps {
    locale: string;
    setLocale: (locale: string) => void;
}

export default function LanguageContextProvider(props: Props) {
    const [locale, setLocale] = useState<string>(props.i18n.resolvedLanguage || props.i18n.language);

    const setLanguagePreference = useCallback(
        (value: string) => {
            setLocale(value);
            props.i18n.changeLanguage(value).then();
        },
        [props.i18n],
    );

    const callback = useCallback(
        (value: string) => {
            setLocalStorage(LocalStorageLanguagePreference, value);
            setLanguagePreference(value);
        },
        [setLanguagePreference],
    );

    const listener = useCallback(
        (ev: StorageEvent): any => {
            if (ev.key !== LocalStorageLanguagePreference) {
                return;
            }

            if (ev.newValue && ev.newValue !== "") {
                setLanguagePreference(ev.newValue);
            }
        },
        [setLanguagePreference],
    );

    useEffect(() => {
        window.addEventListener("storage", listener);

        return () => {
            window.removeEventListener("storage", listener);
        };
    }, [listener]);

    return (
        <LanguageContext.Provider
            value={{
                locale,
                setLocale: callback,
            }}
        >
            {props.children}
        </LanguageContext.Provider>
    );
}

export function useLanguageContext() {
    const context = useContext(LanguageContext);
    if (!context) {
        throw new Error("useLanguageContext must be used within a LanguageContextProvider");
    }

    return context;
}
