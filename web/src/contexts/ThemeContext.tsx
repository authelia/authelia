import React, { createContext, useCallback, useContext, useEffect, useState } from "react";

import { Theme, ThemeProvider } from "@mui/material";

import { LocalStorageThemeName } from "@constants/LocalStorage";
import { localStorageAvailable, setLocalStorage } from "@services/LocalStorage";
import * as themes from "@themes/index";
import { getTheme } from "@utils/Configuration";

const MediaQueryDarkMode = "(prefers-color-scheme: dark)";

export const ThemeContext = createContext<null | ValueProps>(null);

export interface Props {
    children: React.ReactNode;
}

export interface ValueProps {
    theme: Theme;
    themeName: string;
    setThemeName: (value: string) => void;
}

export default function ThemeContextProvider(props: Props) {
    const [theme, setTheme] = useState(GetCurrentTheme());
    const [themeName, setThemeName] = useState(GetCurrentThemeName());

    useEffect(() => {
        if (themeName === themes.ThemeNameAuto) {
            const query = window.matchMedia(MediaQueryDarkMode);
            if (query.addEventListener) {
                query.addEventListener("change", mediaQueryListener);

                return () => {
                    query.removeEventListener("change", mediaQueryListener);
                };
            }
        }

        setTheme(ThemeFromName(themeName));
    }, [themeName]);

    useEffect(() => {
        window.addEventListener("storage", storageListener);

        return () => {
            window.removeEventListener("storage", storageListener);
        };
    }, []);

    const storageListener = (ev: StorageEvent): any => {
        if (ev.key !== LocalStorageThemeName) {
            return;
        }

        if (ev.newValue && ev.newValue !== "") {
            setThemeName(ev.newValue);
        } else {
            setThemeName(getUserThemeName());
        }
    };

    const mediaQueryListener = (ev: MediaQueryListEvent) => {
        setTheme(ev.matches ? themes.Dark : themes.Light);
    };

    const callback = useCallback((name: string) => {
        setThemeName(name);

        setLocalStorage(LocalStorageThemeName, name);
    }, []);

    return (
        <ThemeContext.Provider
            value={{
                setThemeName: callback,
                theme,
                themeName,
            }}
        >
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
        const local = window.localStorage.getItem(LocalStorageThemeName);

        if (local) {
            return local;
        }
    }

    return getTheme();
}

function GetCurrentTheme() {
    return ThemeFromName(GetCurrentThemeName());
}

function ThemeFromName(name: string) {
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
            return window.matchMedia(MediaQueryDarkMode).matches ? themes.Dark : themes.Light;
        default:
            return window.matchMedia(MediaQueryDarkMode).matches ? themes.Dark : themes.Light;
    }
}

const getUserThemeName = () => {
    if (localStorageAvailable()) {
        const value = window.localStorage.getItem(LocalStorageThemeName);

        if (value) {
            return value;
        }
    }

    return getTheme();
};
