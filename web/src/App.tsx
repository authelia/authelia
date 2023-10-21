import React, { Suspense, lazy, useEffect, useState } from "react";

import createCache from "@emotion/cache";
import { CacheProvider } from "@emotion/react";
import { config as faConfig } from "@fortawesome/fontawesome-svg-core";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { Route, BrowserRouter as Router, Routes } from "react-router-dom";

import NotificationBar from "@components/NotificationBar";
import { LocalStorageThemeName } from "@constants/LocalStorage";
import {
    ConsentRoute,
    IndexRoute,
    LogoutRoute,
    ResetPasswordStep1Route,
    ResetPasswordStep2Route,
    SettingsRoute,
} from "@constants/Routes";
import NotificationsContext from "@hooks/NotificationsContext";
import { Notification } from "@models/Notifications";
import { getLocalStorageWithFallback } from "@services/LocalStorage";
import * as themes from "@themes/index";
import { getBasePath } from "@utils/BasePath";
import {
    getDuoSelfEnrollment,
    getRememberMe,
    getResetPassword,
    getResetPasswordCustomURL,
    getTheme,
} from "@utils/Configuration";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import LoginPortal from "@views/LoginPortal/LoginPortal";

import "@fortawesome/fontawesome-svg-core/styles.css";

const ConsentView = lazy(() => import("@views/LoginPortal/ConsentView/ConsentView"));
const SignOut = lazy(() => import("@views/LoginPortal/SignOut/SignOut"));
const ResetPasswordStep1 = lazy(() => import("@views/ResetPassword/ResetPasswordStep1"));
const ResetPasswordStep2 = lazy(() => import("@views/ResetPassword/ResetPasswordStep2"));
const SettingsRouter = lazy(() => import("@views/Settings/SettingsRouter"));

faConfig.autoAddCss = false;

export interface Props {
    nonce?: string;
}

const App: React.FC<Props> = (props: Props) => {
    const [notification, setNotification] = useState(null as Notification | null);
    const [theme, setTheme] = useState(Theme());
    const [themeName, setThemeName] = useState(getUserThemeName());

    const cache = createCache({
        key: "authelia",
        nonce: props.nonce,
        prepend: true,
    });

    useEffect(() => {
        setTheme(ThemeFromName(themeName));
    }, [themeName]);

    useEffect(() => {
        if (getUserThemeName() === themes.ThemeNameAuto) {
            const query = window.matchMedia("(prefers-color-scheme: dark)");
            // MediaQueryLists does not inherit from EventTarget in Internet Explorer
            if (query.addEventListener) {
                const listener = (e: MediaQueryListEvent) => {
                    setTheme(e.matches ? themes.Dark : themes.Light);
                };

                query.addEventListener("change", listener);

                return () => {
                    query.removeEventListener("change", listener);
                };
            }
        }
    }, []);

    useEffect(() => {
        const listener = (e: globalThis.StorageEvent) => {
            if (!e.key || e.key !== LocalStorageThemeName) {
                return;
            }

            if (e.newValue && e.newValue !== "") {
                setThemeName(e.newValue);
            } else {
                setThemeName(getUserThemeName());
            }
        };

        window.addEventListener("storage", listener);

        return () => {
            window.removeEventListener("storage", listener);
        };
    }, []);

    return (
        <CacheProvider value={cache}>
            <ThemeProvider theme={theme}>
                <Suspense fallback={<LoadingPage />}>
                    <CssBaseline />
                    <NotificationsContext.Provider value={{ notification, setNotification }}>
                        <Router basename={getBasePath()}>
                            <NotificationBar onClose={() => setNotification(null)} />
                            <Routes>
                                <Route path={ResetPasswordStep1Route} element={<ResetPasswordStep1 />} />
                                <Route path={ResetPasswordStep2Route} element={<ResetPasswordStep2 />} />
                                <Route path={LogoutRoute} element={<SignOut />} />
                                <Route path={ConsentRoute} element={<ConsentView />} />
                                <Route path={`${SettingsRoute}/*`} element={<SettingsRouter />} />
                                <Route
                                    path={`${IndexRoute}*`}
                                    element={
                                        <LoginPortal
                                            duoSelfEnrollment={getDuoSelfEnrollment()}
                                            rememberMe={getRememberMe()}
                                            resetPassword={getResetPassword()}
                                            resetPasswordCustomURL={getResetPasswordCustomURL()}
                                        />
                                    }
                                />
                            </Routes>
                        </Router>
                    </NotificationsContext.Provider>
                </Suspense>
            </ThemeProvider>
        </CacheProvider>
    );
};

export default App;

function Theme() {
    return ThemeFromName(getUserThemeName());
}

function ThemeFromName(name: string) {
    switch (name) {
        case themes.ThemeNameLight:
            return themes.Light;
        case themes.ThemeNameDark:
            return themes.Dark;
        case themes.ThemeNameGrey:
            return themes.Grey;
        case themes.ThemeNameAuto:
            return window.matchMedia("(prefers-color-scheme: dark)").matches ? themes.Dark : themes.Light;
        default:
            return window.matchMedia("(prefers-color-scheme: dark)").matches ? themes.Dark : themes.Light;
    }
}

const getUserThemeName = () => {
    return getLocalStorageWithFallback(LocalStorageThemeName, getTheme());
};
