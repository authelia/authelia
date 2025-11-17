import { Fragment, ReactNode, lazy, useCallback, useEffect, useState } from "react";

import { Box } from "@mui/material";
import { useTranslation } from "react-i18next";
import { Route, Routes, useLocation } from "react-router-dom";

import {
    AuthenticatedRoute,
    IndexRoute,
    SecondFactorPasswordSubRoute,
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
import LoadingPage from "@views/LoadingPage/LoadingPage";

const AuthenticatedView = lazy(() => import("@views/LoginPortal/AuthenticatedView/AuthenticatedView"));
const FirstFactorForm = lazy(() => import("@views/LoginPortal/FirstFactor/FirstFactorForm"));
const SecondFactorForm = lazy(() => import("@views/LoginPortal/SecondFactor/SecondFactorForm"));

export interface Props {
    duoSelfEnrollment: boolean;
    passkeyLogin: boolean;
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

    useEffect(() => {
        fetchState();
    }, [fetchState]);

    useEffect(() => {
        if (state && state.authentication_level >= AuthenticationLevel.OneFactor) {
            fetchUserInfo();
            fetchConfiguration();
        }
    }, [state, fetchUserInfo, fetchConfiguration]);

    useEffect(() => {
        if (fetchStateError) {
            createErrorNotification(translate("There was an issue retrieving the current user state"));
        }
    }, [fetchStateError, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchConfigurationError) {
            createErrorNotification(translate("There was an issue retrieving global configuration"));
        }
    }, [fetchConfigurationError, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification(translate("There was an issue retrieving user preferences"));
        }
    }, [fetchUserInfoError, createErrorNotification, translate]);

    const handleRedirection = useCallback(async () => {
        if (!redirectionURL) {
            return false;
        }

        const shouldRedirect =
            (configuration?.available_methods?.size === 0 &&
                state!.authentication_level >= AuthenticationLevel.OneFactor) ||
            state!.authentication_level === AuthenticationLevel.TwoFactor ||
            broadcastRedirect;

        if (!shouldRedirect) {
            return false;
        }

        try {
            const res = await checkSafeRedirection(redirectionURL);
            if (res?.ok) {
                redirector(redirectionURL);
            } else {
                createErrorNotification(translate(RedirectionErrorMessage));
            }
        } catch {
            createErrorNotification(translate(RedirectionErrorMessage));
        }

        return true;
    }, [redirectionURL, configuration, state, broadcastRedirect, redirector, createErrorNotification, translate]);

    const handleAuthenticationNavigation = useCallback(() => {
        if (state!.authentication_level === AuthenticationLevel.Unauthenticated) {
            setFirstFactorDisabled(false);
            navigate(IndexRoute);
        } else if (state!.authentication_level >= AuthenticationLevel.OneFactor && userInfo && configuration) {
            if (configuration.available_methods.size === 0) {
                navigate(AuthenticatedRoute, false);
            } else {
                const method = localStorageMethod || userInfo.method;

                if (!state!.factor_knowledge) {
                    navigate(`${SecondFactorRoute}${SecondFactorPasswordSubRoute}`);
                } else if (method === SecondFactorMethod.WebAuthn) {
                    navigate(`${SecondFactorRoute}${SecondFactorWebAuthnSubRoute}`);
                } else if (method === SecondFactorMethod.MobilePush) {
                    navigate(`${SecondFactorRoute}${SecondFactorPushSubRoute}`);
                } else {
                    navigate(`${SecondFactorRoute}${SecondFactorTOTPSubRoute}`);
                }
            }
        }
    }, [state, userInfo, configuration, navigate, localStorageMethod]);

    useEffect(() => {
        (async function () {
            if (!state) {
                return;
            }

            const redirected = await handleRedirection();
            if (redirected) {
                return;
            }

            handleAuthenticationNavigation();
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
        handleAuthenticationNavigation,
        handleRedirection,
    ]);

    const handleChannelStateChange = async () => {
        setBroadcastRedirect(true);
        fetchState();
    };

    const handleAuthSuccess = async (redirectionURL: string | undefined) => {
        if (redirectionURL) {
            redirector(redirectionURL);
        } else {
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
                            passkeyLogin={props.passkeyLogin}
                            rememberMe={props.rememberMe}
                            resetPassword={props.resetPassword}
                            resetPasswordCustomURL={props.resetPasswordCustomURL}
                            onAuthenticationStart={() => setFirstFactorDisabled(true)}
                            onAuthenticationStop={() => setFirstFactorDisabled(false)}
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
                            factorKnowledge={state.factor_knowledge}
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

interface ComponentOrLoadingProps {
    readonly ready: boolean;
    readonly children: ReactNode;
}

function ComponentOrLoading(props: ComponentOrLoadingProps) {
    return (
        <Fragment>
            <Box className={props.ready ? "hidden" : ""}>
                <LoadingPage />
            </Box>
            {props.ready ? props.children : null}
        </Fragment>
    );
}

export default LoginPortal;
