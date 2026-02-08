import { ReactNode, createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

import { i18n } from "i18next";

import { LocalStorageLanguagePreference } from "@constants/LocalStorage";
import { setLocalStorage } from "@services/LocalStorage";

export const LanguageContext = createContext<null | ValueProps>(null);

export interface Props {
    readonly i18n: i18n;
    readonly children: ReactNode;
}

export interface ValueProps {
    readonly locale: string;
    readonly setLocale: (_locale: string) => void;
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
        globalThis.addEventListener("storage", listener);

        return () => {
            globalThis.removeEventListener("storage", listener);
        };
    }, [listener]);

    return (
        <LanguageContext.Provider
            value={useMemo(
                () => ({
                    locale,
                    setLocale: callback,
                }),
                [locale, callback],
            )}
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
