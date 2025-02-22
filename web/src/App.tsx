import React, { Suspense, lazy, useState } from "react";

import createCache from "@emotion/cache";
import { CacheProvider } from "@emotion/react";
import { config as faConfig } from "@fortawesome/fontawesome-svg-core";
import { CssBaseline } from "@mui/material";
import { Route, BrowserRouter as Router, Routes } from "react-router-dom";

import NotificationBar from "@components/NotificationBar";
import {
    ConsentRoute,
    ErrorRoute,
    IndexRoute,
    LogoutRoute,
    ResetPasswordStep1Route,
    ResetPasswordStep2Route,
    RevokeOneTimeCodeRoute,
    RevokeResetPasswordRoute,
    SettingsRoute,
} from "@constants/Routes";
import LocalStorageMethodContextProvider from "@contexts/LocalStorageMethodContext";
import ThemeContextProvider from "@contexts/ThemeContext";
import NotificationsContext from "@hooks/NotificationsContext";
import { Notification } from "@models/Notifications";
import { getBasePath } from "@utils/BasePath";
import { getDuoSelfEnrollment, getRememberMe, getResetPassword, getResetPasswordCustomURL } from "@utils/Configuration";
import BaseErrorView from "@views/Error/BaseErrorPage";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import LoginPortal from "@views/LoginPortal/LoginPortal";

import "@fortawesome/fontawesome-svg-core/styles.css";

const ConsentView = lazy(() => import("@views/LoginPortal/ConsentView/ConsentView"));
const SignOut = lazy(() => import("@views/LoginPortal/SignOut/SignOut"));
const ResetPasswordStep1 = lazy(() => import("@views/ResetPassword/ResetPasswordStep1"));
const ResetPasswordStep2 = lazy(() => import("@views/ResetPassword/ResetPasswordStep2"));
const SettingsRouter = lazy(() => import("@views/Settings/SettingsRouter"));
const RevokeOneTimeCodeView = lazy(() => import("@views/Revoke/RevokeOneTimeCodeView"));
const RevokeResetPasswordTokenView = lazy(() => import("@views/Revoke/RevokeResetPasswordTokenView"));

faConfig.autoAddCss = false;

export interface Props {
    nonce?: string;
}

const App: React.FC<Props> = (props: Props) => {
    const [notification, setNotification] = useState(null as Notification | null);

    const cache = createCache({
        key: "authelia",
        nonce: props.nonce,
        prepend: true,
    });

    return (
        <CacheProvider value={cache}>
            <ThemeContextProvider>
                <Suspense fallback={<LoadingPage />}>
                    <CssBaseline />
                    <NotificationsContext.Provider value={{ notification, setNotification }}>
                        <LocalStorageMethodContextProvider>
                            <Router basename={getBasePath()}>
                                <NotificationBar onClose={() => setNotification(null)} />
                                <Routes>
                                    <Route path={ResetPasswordStep1Route} element={<ResetPasswordStep1 />} />
                                    <Route path={ResetPasswordStep2Route} element={<ResetPasswordStep2 />} />
                                    <Route path={LogoutRoute} element={<SignOut />} />
                                    <Route path={ConsentRoute} element={<ConsentView />} />
                                    <Route path={ErrorRoute} element={<BaseErrorView />} />
                                    <Route path={RevokeOneTimeCodeRoute} element={<RevokeOneTimeCodeView />} />
                                    <Route path={RevokeResetPasswordRoute} element={<RevokeResetPasswordTokenView />} />
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
                        </LocalStorageMethodContextProvider>
                    </NotificationsContext.Provider>
                </Suspense>
            </ThemeContextProvider>
        </CacheProvider>
    );
};

export default App;
