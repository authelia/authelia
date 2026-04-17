import {
    ReactNode,
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useState,
    useSyncExternalStore,
} from "react";

import { Theme, ThemeProvider } from "@mui/material";

import { LocalStorageThemeName } from "@constants/LocalStorage";
import { localStorageAvailable, setLocalStorage } from "@services/LocalStorage";
import * as themes from "@themes/index";
import { getTheme } from "@utils/Configuration";

const MediaQueryDarkMode = "(prefers-color-scheme: dark)";

export const ThemeContext = createContext<null | ValueProps>(null);

export interface Props {
    readonly children: ReactNode;
}

export interface ValueProps {
    theme: Theme;
    themeName: string;
    setThemeName: (_value: string) => void;
}

export default function ThemeContextProvider(props: Props) {
    const [themeName, setThemeName] = useState(GetCurrentThemeName());
    const prefersDark = useSyncExternalStore(subscribePrefersDark, getPrefersDarkSnapshot, getPrefersDarkSnapshot);

    const theme = useMemo(() => ThemeFromName(themeName, prefersDark), [themeName, prefersDark]);

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
            theme,
            themeName,
        }),
        [callback, theme, themeName],
    );

    return (
        <ThemeContext.Provider value={value}>
            <ThemeWrapper>{props.children}</ThemeWrapper>
        </ThemeContext.Provider>
    );
}

export function useThemeContext() {
    const context = useContext(ThemeContext);
    if (!context) {
        throw new Error("useThemeContext must be used within a ThemeContextProvider");
    }

    return context;
}

function ThemeWrapper(props: Props) {
    const { theme } = useThemeContext();

    return <ThemeProvider theme={theme}>{props.children}</ThemeProvider>;
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

function ThemeFromName(name: string, prefersDark: boolean) {
    switch (name) {
        case themes.ThemeNameLight:
            return themes.Light;
        case themes.ThemeNameDark:
            return themes.Dark;
        case themes.ThemeNameGrey:
            return themes.Grey;
        case themes.ThemeNameOled:
            return themes.Oled;
        case themes.ThemeNameAuto:
        default:
            return prefersDark ? themes.Dark : themes.Light;
    }
}
