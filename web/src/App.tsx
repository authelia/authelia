// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Suspense, lazy } from "react";

import { CSPProvider } from "@base-ui/react/csp-provider";
import { useTranslation } from "react-i18next";
import { Route, BrowserRouter as Router, Routes } from "react-router";

import { TooltipProvider } from "@components/UI/Tooltip";
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
import LanguageContextProvider from "@contexts/LanguageContext";
import LocalStorageMethodContextProvider from "@contexts/LocalStorageMethodContext";
import NotificationsContextProvider from "@contexts/NotificationsContext";
import ThemeContextProvider from "@contexts/ThemeContext";
import { getBasePath } from "@utils/BasePath";
import {
    getCSPNonce,
    getDuoSelfEnrollment,
    getPasskeyLogin,
    getRegistrationURL,
    getRememberMe,
    getResetPassword,
    getResetPasswordCustomURL,
} from "@utils/Configuration";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import LoginPortal from "@views/LoginPortal/LoginPortal";

const ConsentPortal = lazy(() => import("@views/ConsentPortal/ConsentPortal"));
const ErrorView = lazy(() => import("@views/Error/ErrorView"));
const ResetPasswordStep1 = lazy(() => import("@views/ResetPassword/ResetPasswordStep1"));
const ResetPasswordStep2 = lazy(() => import("@views/ResetPassword/ResetPasswordStep2"));
const RevokeOneTimeCodeView = lazy(() => import("@views/Revoke/RevokeOneTimeCodeView"));
const RevokeResetPasswordTokenView = lazy(() => import("@views/Revoke/RevokeResetPasswordTokenView"));
const SettingsRouter = lazy(() => import("@views/Settings/SettingsRouter"));
const SignOut = lazy(() => import("@views/LoginPortal/SignOut/SignOut"));

function App() {
    const { i18n } = useTranslation();

    return (
        <CSPProvider nonce={getCSPNonce()}>
            <LanguageContextProvider i18n={i18n}>
                <ThemeContextProvider>
                    <Suspense fallback={<LoadingPage />}>
                        <TooltipProvider>
                            <NotificationsContextProvider>
                                <LocalStorageMethodContextProvider>
                                    <Router basename={getBasePath()}>
                                        <Routes>
                                            <Route path={`${ConsentRoute}/*`} element={<ConsentPortal />} />
                                            <Route path={ErrorRoute} element={<ErrorView />} />
                                            <Route
                                                path={`${IndexRoute}*`}
                                                element={
                                                    <LoginPortal
                                                        duoSelfEnrollment={getDuoSelfEnrollment()}
                                                        passkeyLogin={getPasskeyLogin()}
                                                        rememberMe={getRememberMe()}
                                                        resetPassword={getResetPassword()}
                                                        resetPasswordCustomURL={getResetPasswordCustomURL()}
                                                        registrationURL={getRegistrationURL()}
                                                    />
                                                }
                                            />
                                            <Route path={LogoutRoute} element={<SignOut />} />
                                            <Route path={ResetPasswordStep1Route} element={<ResetPasswordStep1 />} />
                                            <Route path={ResetPasswordStep2Route} element={<ResetPasswordStep2 />} />
                                            <Route path={RevokeOneTimeCodeRoute} element={<RevokeOneTimeCodeView />} />
                                            <Route
                                                path={RevokeResetPasswordRoute}
                                                element={<RevokeResetPasswordTokenView />}
                                            />
                                            <Route path={`${SettingsRoute}/*`} element={<SettingsRouter />} />
                                        </Routes>
                                    </Router>
                                </LocalStorageMethodContextProvider>
                            </NotificationsContextProvider>
                        </TooltipProvider>
                    </Suspense>
                </ThemeContextProvider>
            </LanguageContextProvider>
        </CSPProvider>
    );
}

export default App;
