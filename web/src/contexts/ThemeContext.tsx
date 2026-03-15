import { ReactNode, createContext, use, useCallback, useEffect, useMemo, useState, useSyncExternalStore } from "react";

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
    const [themeName, setThemeName] = useState(() => GetCurrentThemeName());
    const prefersDark = useSyncExternalStore(subscribePrefersDark, getPrefersDarkSnapshot, getPrefersDarkSnapshot);

    useEffect(() => {
        document.documentElement.setAttribute("data-theme", ResolveThemeName(themeName, prefersDark));
    }, [themeName, prefersDark]);

    useEffect(() => {
        const listener = (ev: StorageEvent) => {
            if (ev.key !== LocalStorageThemeName) {
                return;
            }

            if (ev.newValue && ev.newValue !== "") {
                setThemeName(ev.newValue);
            } else {
                setThemeName(GetCurrentThemeName());
            }
        };

        globalThis.addEventListener?.("storage", listener);

        return () => {
            globalThis.removeEventListener?.("storage", listener);
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

    return <ThemeContext value={value}>{props.children}</ThemeContext>;
}

export function useThemeContext() {
    const context = use(ThemeContext);
    if (!context) {
        throw new Error("useThemeContext must be used within a ThemeContextProvider");
    }

    return context;
}

function ResolveThemeName(name: string, prefersDark: boolean): string {
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
        default:
            return prefersDark ? "dark" : "light";
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

function subscribePrefersDark(listener: () => void): () => void {
    const query = globalThis.matchMedia?.(MediaQueryDarkMode);
    if (!query?.addEventListener) {
        return () => {};
    }
    query.addEventListener("change", listener);
    return () => query.removeEventListener("change", listener);
}

function getPrefersDarkSnapshot(): boolean {
    return globalThis.matchMedia?.(MediaQueryDarkMode).matches ?? false;
}
