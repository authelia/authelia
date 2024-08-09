import React, { lazy, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";
import { Route, Routes, useLocation } from "react-router-dom";

import {
    AuthenticatedRoute,
    IndexRoute,
    SecondFactorPushSubRoute,
    SecondFactorRoute,
    SecondFactorTOTPSubRoute,
    SecondFactorWebAuthnSubRoute,
} from "@constants/Routes";
import { RedirectionURL } from "@constants/SearchParams";
import { useLocalStorageMethodContext } from "@contexts/LocalStorageMethodContext";
import { useConfiguration } from "@hooks/Configuration";
import { useNotifications } from "@hooks/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useRedirector } from "@hooks/Redirector";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import { useUserInfoPOST } from "@hooks/UserInfo";
import { SecondFactorMethod } from "@models/Methods";
import { checkSafeRedirection } from "@services/SafeRedirection";
import { AuthenticationLevel } from "@services/State";
import { ComponentOrLoading } from "@views/Generic/ComponentOrLoading";

const AuthenticatedView = lazy(() => import("@views/LoginPortal/AuthenticatedView/AuthenticatedView"));
const FirstFactorForm = lazy(() => import("@views/LoginPortal/FirstFactor/FirstFactorForm"));
const SecondFactorForm = lazy(() => import("@views/LoginPortal/SecondFactor/SecondFactorForm"));

export interface Props {
    duoSelfEnrollment: boolean;
    rememberMe: boolean;

    resetPassword: boolean;
    resetPasswordCustomURL: string;
}

const RedirectionErrorMessage =
    "Redirection was determined to be unsafe and aborted ensure the redirection URL is correct";

const LoginPortal = function (props: Props) {
    const location = useLocation();
    const redirectionURL = useQueryParam(RedirectionURL);
    const { createErrorNotification } = useNotifications();
    const [firstFactorDisabled, setFirstFactorDisabled] = useState(true);
    const [broadcastRedirect, setBroadcastRedirect] = useState(false);
    const redirector = useRedirector();
    const { localStorageMethod } = useLocalStorageMethodContext();
    const { t: translate } = useTranslation();

    const [state, fetchState, , fetchStateError] = useAutheliaState();
    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoPOST();
    const [configuration, fetchConfiguration, , fetchConfigurationError] = useConfiguration();

    const navigate = useRouterNavigate();

    // Fetch the state when portal is mounted.
    useEffect(() => {
        fetchState();
    }, [fetchState]);

    // Fetch preferences and configuration when user is authenticated.
    useEffect(() => {
        if (state && state.authentication_level >= AuthenticationLevel.OneFactor) {
            fetchUserInfo();
            fetchConfiguration();
        }
    }, [state, fetchUserInfo, fetchConfiguration]);

    // Enable first factor when user is unauthenticated.
    useEffect(() => {
        if (state && state.authentication_level > AuthenticationLevel.Unauthenticated) {
            setFirstFactorDisabled(true);
        }
    }, [state, setFirstFactorDisabled]);

    // Display an error when state fetching fails
    useEffect(() => {
        if (fetchStateError) {
            createErrorNotification(translate("There was an issue retrieving the current user state"));
        }
    }, [fetchStateError, createErrorNotification, translate]);

    // Display an error when configuration fetching fails
    useEffect(() => {
        if (fetchConfigurationError) {
            createErrorNotification(translate("There was an issue retrieving global configuration"));
        }
    }, [fetchConfigurationError, createErrorNotification, translate]);

    // Display an error when preferences fetching fails
    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification(translate("There was an issue retrieving user preferences"));
        }
    }, [fetchUserInfoError, createErrorNotification, translate]);

    // Redirect to the correct stage if not enough authenticated
    useEffect(() => {
        (async function () {
            if (!state) {
                return;
            }

            if (
                redirectionURL &&
                ((configuration &&
                    configuration.available_methods.size === 0 &&
                    state.authentication_level >= AuthenticationLevel.OneFactor) ||
                    state.authentication_level === AuthenticationLevel.TwoFactor ||
                    broadcastRedirect)
            ) {
                try {
                    const res = await checkSafeRedirection(redirectionURL);
                    if (res && res.ok) {
                        redirector(redirectionURL);
                    } else {
                        createErrorNotification(translate(RedirectionErrorMessage));
                    }
                } catch (err) {
                    createErrorNotification(translate(RedirectionErrorMessage));
                }

                return;
            }

            if (state.authentication_level === AuthenticationLevel.Unauthenticated) {
                setFirstFactorDisabled(false);
                navigate(IndexRoute);
            } else if (state.authentication_level >= AuthenticationLevel.OneFactor && userInfo && configuration) {
                if (configuration.available_methods.size === 0) {
                    navigate(AuthenticatedRoute, false);
                } else {
                    const method = localStorageMethod || userInfo.method;

                    if (method === SecondFactorMethod.WebAuthn) {
                        navigate(`${SecondFactorRoute}${SecondFactorWebAuthnSubRoute}`);
                    } else if (method === SecondFactorMethod.MobilePush) {
                        navigate(`${SecondFactorRoute}${SecondFactorPushSubRoute}`);
                    } else {
                        navigate(`${SecondFactorRoute}${SecondFactorTOTPSubRoute}`);
                    }
                }
            }
        })();
    }, [
        state,
        redirectionURL,
        navigate,
        userInfo,
        setFirstFactorDisabled,
        configuration,
        createErrorNotification,
        redirector,
        broadcastRedirect,
        localStorageMethod,
        translate,
    ]);

    const handleChannelStateChange = async () => {
        setBroadcastRedirect(true);
        fetchState();
    };

    const handleAuthSuccess = async (redirectionURL: string | undefined) => {
        if (redirectionURL) {
            // Do an external redirection pushed by the server.
            redirector(redirectionURL);
        } else {
            // Refresh state
            fetchState();
        }
    };

    const firstFactorReady =
        state !== undefined &&
        state.authentication_level === AuthenticationLevel.Unauthenticated &&
        location.pathname === IndexRoute;

    return (
        <Routes>
            <Route
                path={IndexRoute}
                element={
                    <ComponentOrLoading ready={firstFactorReady}>
                        <FirstFactorForm
                            disabled={firstFactorDisabled}
                            rememberMe={props.rememberMe}
                            resetPassword={props.resetPassword}
                            resetPasswordCustomURL={props.resetPasswordCustomURL}
                            onAuthenticationStart={() => setFirstFactorDisabled(true)}
                            onAuthenticationFailure={() => setFirstFactorDisabled(false)}
                            onAuthenticationSuccess={handleAuthSuccess}
                            onChannelStateChange={handleChannelStateChange}
                        />
                    </ComponentOrLoading>
                }
            />
            <Route
                path={`${SecondFactorRoute}/*`}
                element={
                    state && userInfo && configuration ? (
                        <SecondFactorForm
                            authenticationLevel={state.authentication_level}
                            userInfo={userInfo}
                            configuration={configuration}
                            duoSelfEnrollment={props.duoSelfEnrollment}
                            onMethodChanged={() => fetchUserInfo()}
                            onAuthenticationSuccess={handleAuthSuccess}
                        />
                    ) : null
                }
            />
            <Route path={AuthenticatedRoute} element={userInfo ? <AuthenticatedView userInfo={userInfo} /> : null} />
        </Routes>
    );
};

export default LoginPortal;
