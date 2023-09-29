import React, { Suspense, useEffect, useState } from "react";

import createCache from "@emotion/cache";
import { CacheProvider } from "@emotion/react";
import { config as faConfig } from "@fortawesome/fontawesome-svg-core";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { Route, BrowserRouter as Router, Routes } from "react-router-dom";

import NotificationBar from "@components/NotificationBar";
import {
    ConsentRoute,
    IndexRoute,
    LogoutRoute,
    RegisterOneTimePasswordRoute,
    ResetPasswordStep1Route,
    ResetPasswordStep2Route,
    SettingsRoute,
} from "@constants/Routes";
import NotificationsContext from "@hooks/NotificationsContext";
import { Notification } from "@models/Notifications";
import * as themes from "@themes/index";
import { getBasePath } from "@utils/BasePath";
import {
    getDuoSelfEnrollment,
    getRememberMe,
    getResetPassword,
    getResetPasswordCustomURL,
    getTheme,
} from "@utils/Configuration";
import RegisterOneTimePassword from "@views/DeviceRegistration/RegisterOneTimePassword";
import BaseLoadingPage from "@views/LoadingPage/BaseLoadingPage";
import ConsentView from "@views/LoginPortal/ConsentView/ConsentView";
import LoginPortal from "@views/LoginPortal/LoginPortal";
import SignOut from "@views/LoginPortal/SignOut/SignOut";
import ResetPasswordStep1 from "@views/ResetPassword/ResetPasswordStep1";
import ResetPasswordStep2 from "@views/ResetPassword/ResetPasswordStep2";
import SettingsRouter from "@views/Settings/SettingsRouter";

import "@fortawesome/fontawesome-svg-core/styles.css";

faConfig.autoAddCss = false;

function Theme() {
    switch (getTheme()) {
        case "dark":
            return themes.Dark;
        case "grey":
            return themes.Grey;
        case "auto":
            return window.matchMedia("(prefers-color-scheme: dark)").matches ? themes.Dark : themes.Light;
        default:
            return themes.Light;
    }
}

export interface Props {
    nonce?: string;
}

const App: React.FC<Props> = (props: Props) => {
    const [notification, setNotification] = useState(null as Notification | null);
    const [theme, setTheme] = useState(Theme());

    const cache = createCache({
        key: "authelia",
        nonce: props.nonce,
        prepend: true,
    });

    useEffect(() => {
        if (getTheme() === "auto") {
            const query = window.matchMedia("(prefers-color-scheme: dark)");
            // MediaQueryLists does not inherit from EventTarget in Internet Explorer
            if (query.addEventListener) {
                query.addEventListener("change", (e) => {
                    setTheme(e.matches ? themes.Dark : themes.Light);
                });
            }
        }
    }, []);
    return (
        <CacheProvider value={cache}>
            <ThemeProvider theme={theme}>
                <Suspense fallback={<BaseLoadingPage message={"Loading"} />}>
                    <CssBaseline />
                    <NotificationsContext.Provider value={{ notification, setNotification }}>
                        <Router basename={getBasePath()}>
                            <NotificationBar onClose={() => setNotification(null)} />
                            <Routes>
                                <Route path={ResetPasswordStep1Route} element={<ResetPasswordStep1 />} />
                                <Route path={ResetPasswordStep2Route} element={<ResetPasswordStep2 />} />
                                <Route path={RegisterOneTimePasswordRoute} element={<RegisterOneTimePassword />} />
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
