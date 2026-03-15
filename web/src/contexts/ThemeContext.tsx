import { ReactNode, createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

import { LocalStorageThemeName } from "@constants/LocalStorage";
import { localStorageAvailable, setLocalStorage } from "@services/LocalStorage";
import { ThemeNameAuto, ThemeNameDark, ThemeNameGrey, ThemeNameLight, ThemeNameOled } from "@themes/index";
import { getTheme } from "@utils/Configuration";

const MediaQueryDarkMode = "(prefers-color-scheme: dark)";

export const ThemeContext = createContext<null | ValueProps>(null);

export interface Props {
    readonly children: ReactNode;
}

export interface ValueProps {
    themeName: string;
    setThemeName: (_value: string) => void;
}

export default function ThemeContextProvider(props: Props) {
    const [themeName, setThemeName] = useState(GetCurrentThemeName());

    useEffect(() => {
        document.documentElement.setAttribute("data-theme", ResolveThemeName(themeName));

        if (themeName === ThemeNameAuto) {
            const query = globalThis.matchMedia?.(MediaQueryDarkMode);
            if (query?.addEventListener) {
                const listener = (ev: MediaQueryListEvent) => {
                    document.documentElement.setAttribute("data-theme", ev.matches ? "dark" : "light");
                };

                query.addEventListener("change", listener);

                return () => {
                    query.removeEventListener("change", listener);
                };
            }
        }
    }, [themeName]);

    useEffect(() => {
        const storageListener = (ev: StorageEvent) => {
            if (ev.key !== LocalStorageThemeName) {
                return;
            }

            if (ev.newValue && ev.newValue !== "") {
                setThemeName(ev.newValue);
            } else {
                setThemeName(getUserThemeName());
            }
        };

        globalThis.addEventListener?.("storage", storageListener);

        return () => {
            globalThis.removeEventListener?.("storage", storageListener);
        };
    }, []);

    const callback = useCallback((name: string) => {
        setThemeName(name);

        setLocalStorage(LocalStorageThemeName, name);
    }, []);

    const value = useMemo(
        () => ({
            setThemeName: callback,
            themeName,
        }),
        [callback, themeName],
    );

    return <ThemeContext.Provider value={value}>{props.children}</ThemeContext.Provider>;
}

export function useThemeContext() {
    const context = useContext(ThemeContext);
    if (!context) {
        throw new Error("useThemeContext must be used within a ThemeContextProvider");
    }

    return context;
}

function ResolveThemeName(name: string): string {
    switch (name) {
        case ThemeNameLight:
            return "light";
        case ThemeNameDark:
            return "dark";
        case ThemeNameGrey:
            return "grey";
        case ThemeNameOled:
            return "oled";
        case ThemeNameAuto:
            return globalThis.matchMedia?.(MediaQueryDarkMode).matches ? "dark" : "light";
        default:
            return globalThis.matchMedia?.(MediaQueryDarkMode).matches ? "dark" : "light";
    }
}

function GetCurrentThemeName() {
    if (localStorageAvailable()) {
        const local = globalThis.localStorage?.getItem(LocalStorageThemeName);

        if (local) {
            return local;
        }
    }

    return getTheme();
}

const getUserThemeName = () => {
    if (localStorageAvailable()) {
        const value = globalThis.localStorage?.getItem(LocalStorageThemeName);

        if (value) {
            return value;
        }
    }

    return getTheme();
};
