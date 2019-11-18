import React, { useEffect, Fragment, ReactNode, useState } from "react";
import { Switch, Route, Redirect, useHistory, useLocation } from "react-router";
import FirstFactorForm from "./FirstFactor/FirstFactorForm";
import SecondFactorForm from "./SecondFactor/SecondFactorForm";
import { FirstFactorRoute, SecondFactorRoute, SecondFactorTOTPRoute, SecondFactorPushRoute, SecondFactorU2FRoute, LogoutRoute } from "../../Routes";
import { useAutheliaState } from "../../hooks/State";
import LoadingPage from "../LoadingPage/LoadingPage";
import { AuthenticationLevel } from "../../services/State";
import { useNotifications } from "../../hooks/NotificationsContext";
import { useRedirectionURL } from "../../hooks/RedirectionURL";
import { useUserPreferences } from "../../hooks/UserPreferences";
import { SecondFactorMethod } from "../../models/Methods";
import { useAutheliaConfiguration } from "../../hooks/Configuration";
import SignOut from "./SignOut/SignOut";

export default function () {
    const history = useHistory();
    const location = useLocation();
    const redirectionURL = useRedirectionURL();
    const { createErrorNotification } = useNotifications();
    const [firstFactorDisabled, setFirstFactorDisabled] = useState(true);

    const [state, fetchState, , fetchStateError] = useAutheliaState();
    const [preferences, fetchPreferences, , fetchPreferencesError] = useUserPreferences();
    const [configuration, fetchConfiguration, , fetchConfigurationError] = useAutheliaConfiguration();

    // Fetch the state when portal is mounted.
    useEffect(() => { fetchState() }, [fetchState]);

    // Fetch preferences and configuration when user is authenticated.
    useEffect(() => {
        if (state && state.authentication_level >= AuthenticationLevel.OneFactor) {
            fetchPreferences();
            fetchConfiguration();
        }
    }, [state, fetchPreferences, fetchConfiguration]);

    // Enable first factor when user is unauthenticated.
    useEffect(() => {
        if (state && state.authentication_level > AuthenticationLevel.Unauthenticated) {
            setFirstFactorDisabled(true);
        }
    }, [state, setFirstFactorDisabled]);

    // Display an error when state fetching fails
    useEffect(() => {
        if (fetchStateError) {
            createErrorNotification("There was an issue fetching the current user state");
        }
    }, [fetchStateError, createErrorNotification]);

    // Display an error when configuration fetching fails
    useEffect(() => {
        if (fetchConfigurationError) {
            createErrorNotification("There was an issue retrieving global configuration");
        }
    }, [fetchConfigurationError, createErrorNotification]);

    // Display an error when preferences fetching fails
    useEffect(() => {
        if (fetchPreferencesError) {
            createErrorNotification("There was an issue retrieving user preferences");
        }
    }, [fetchPreferencesError, createErrorNotification]);

    // Redirect to the correct stage if not enough authenticated
    useEffect(() => {
        if (state) {
            const redirectionSuffix = redirectionURL
                ? `?rd=${encodeURI(redirectionURL)}`
                : '';

            if (state.authentication_level === AuthenticationLevel.Unauthenticated) {
                setFirstFactorDisabled(false);
                history.push(`${FirstFactorRoute}${redirectionSuffix}`);
            } else if (state.authentication_level >= AuthenticationLevel.OneFactor && preferences) {
                console.log("redirect");
                if (preferences.method === SecondFactorMethod.U2F) {
                    history.push(`${SecondFactorU2FRoute}${redirectionSuffix}`);
                } else if (preferences.method === SecondFactorMethod.Duo) {
                    history.push(`${SecondFactorPushRoute}${redirectionSuffix}`);
                } else {
                    history.push(`${SecondFactorTOTPRoute}${redirectionSuffix}`);
                }
            }
        }
    }, [state, redirectionURL, history.push, preferences, setFirstFactorDisabled]);

    const handleFirstFactorSuccess = async (redirectionURL: string | undefined) => {
        if (redirectionURL) {
            // Do an external redirection pushed by the server.
            window.location.href = redirectionURL;
        } else {
            // Refresh state
            fetchState();
        }
    }

    const handleSecondFactorSuccess = async (redirectionURL: string | undefined) => {
        if (redirectionURL) {
            // Do an external redirection pushed by the server.
            window.location.href = redirectionURL;
        } else {
            fetchState();
        }
    }

    const firstFactorReady = state !== undefined &&
        state.authentication_level === AuthenticationLevel.Unauthenticated &&
        location.pathname === FirstFactorRoute;

    return (
        <Switch>
            <Route path={FirstFactorRoute} exact>
                <ComponentOrLoading ready={firstFactorReady}>
                    <FirstFactorForm
                        disabled={firstFactorDisabled}
                        onAuthenticationStart={() => setFirstFactorDisabled(true)}
                        onAuthenticationFailure={() => setFirstFactorDisabled(false)}
                        onAuthenticationSuccess={handleFirstFactorSuccess} />
                </ComponentOrLoading>
            </Route>
            <Route path={SecondFactorRoute}>
                {state && preferences && configuration ? <SecondFactorForm
                    username={state.username}
                    authenticationLevel={state.authentication_level}
                    userPreferences={preferences}
                    configuration={configuration}
                    onMethodChanged={() => fetchPreferences()}
                    onAuthenticationSuccess={handleSecondFactorSuccess} /> : null}
            </Route>
            <Route path="/">
                <Redirect to={FirstFactorRoute} />
            </Route>
        </Switch>
    )
}

interface ComponentOrLoadingProps {
    ready: boolean;

    children: ReactNode;
}

function ComponentOrLoading(props: ComponentOrLoadingProps) {
    return (
        <Fragment>
            <div className={props.ready ? "hidden" : ""}>
                <LoadingPage />
            </div>
            {props.ready ? props.children : null}
        </Fragment>
    )
}