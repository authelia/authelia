import { Suspense, lazy, useMemo, useState } from "react";

import { config as faConfig } from "@fortawesome/fontawesome-svg-core";
import { CssBaseline } from "@mui/material";
import { useTranslation } from "react-i18next";
import { Route, BrowserRouter as Router, Routes } from "react-router-dom";

import NotificationBar from "@components/NotificationBar";
import {
    ConsentRoute,
    IndexRoute,
    LogoutRoute,
    ResetPasswordStep1Route,
    ResetPasswordStep2Route,
    RevokeOneTimeCodeRoute,
    RevokeResetPasswordRoute,
    SettingsRoute,
} from "@constants/Routes";
import LanguageContextProvider from "@contexts/LanguageContext";
import LocalStorageMethodContextProvider from "@contexts/LocalStorageMethodContext";
import ThemeContextProvider from "@contexts/ThemeContext";
import NotificationsContext from "@hooks/NotificationsContext";
import { Notification } from "@models/Notifications";
import { getBasePath } from "@utils/BasePath";
import {
    getDuoSelfEnrollment,
    getPasskeyLogin,
    getRememberMe,
    getResetPassword,
    getResetPasswordCustomURL,
} from "@utils/Configuration";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import LoginPortal from "@views/LoginPortal/LoginPortal";

import "@fortawesome/fontawesome-svg-core/styles.css";

const ConsentPortal = lazy(() => import("@views/ConsentPortal/ConsentPortal"));
const SignOut = lazy(() => import("@views/LoginPortal/SignOut/SignOut"));
const ResetPasswordStep1 = lazy(() => import("@views/ResetPassword/ResetPasswordStep1"));
const ResetPasswordStep2 = lazy(() => import("@views/ResetPassword/ResetPasswordStep2"));
const SettingsRouter = lazy(() => import("@views/Settings/SettingsRouter"));
const RevokeOneTimeCodeView = lazy(() => import("@views/Revoke/RevokeOneTimeCodeView"));
const RevokeResetPasswordTokenView = lazy(() => import("@views/Revoke/RevokeResetPasswordTokenView"));

faConfig.autoAddCss = false;

function App() {
    const [notification, setNotification] = useState(null as Notification | null);
    const { i18n } = useTranslation();

    const notificationsContextValue = useMemo(() => ({ notification, setNotification }), [notification]);

    return (
        <LanguageContextProvider i18n={i18n}>
            <ThemeContextProvider>
                <Suspense fallback={<LoadingPage />}>
                    <CssBaseline />
                    <NotificationsContext.Provider value={notificationsContextValue}>
                        <LocalStorageMethodContextProvider>
                            <Router basename={getBasePath()}>
                                <NotificationBar onClose={() => setNotification(null)} />
                                <Routes>
                                    <Route path={ResetPasswordStep1Route} element={<ResetPasswordStep1 />} />
                                    <Route path={ResetPasswordStep2Route} element={<ResetPasswordStep2 />} />
                                    <Route path={LogoutRoute} element={<SignOut />} />
                                    <Route path={RevokeOneTimeCodeRoute} element={<RevokeOneTimeCodeView />} />
                                    <Route path={RevokeResetPasswordRoute} element={<RevokeResetPasswordTokenView />} />
                                    <Route path={`${SettingsRoute}/*`} element={<SettingsRouter />} />
                                    <Route path={`${ConsentRoute}/*`} element={<ConsentPortal />} />
                                    <Route
                                        path={`${IndexRoute}*`}
                                        element={
                                            <LoginPortal
                                                duoSelfEnrollment={getDuoSelfEnrollment()}
                                                passkeyLogin={getPasskeyLogin()}
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
        </LanguageContextProvider>
    );
}

export default App;
