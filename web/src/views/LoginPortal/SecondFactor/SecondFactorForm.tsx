import React, { lazy, useEffect, useState } from "react";

import { Button, Grid, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { browserSupportsWebAuthn } from "@simplewebauthn/browser";
import { useTranslation } from "react-i18next";
import { Route, Routes, useNavigate } from "react-router-dom";

import {
    SecondFactorPushSubRoute,
    SecondFactorTOTPSubRoute,
    SecondFactorWebAuthnSubRoute,
    SettingsRoute,
    SettingsTwoFactorAuthenticationSubRoute,
    LogoutRoute as SignOutRoute,
} from "@constants/Routes";
import { useLocalStorageMethodContext } from "@contexts/LocalStorageMethodContext";
import { useNotifications } from "@hooks/NotificationsContext";
import LoginLayout from "@layouts/LoginLayout";
import { Configuration } from "@models/Configuration";
import { SecondFactorMethod } from "@models/Methods";
import { UserInfo } from "@models/UserInfo";
import { AuthenticationLevel } from "@services/State";
import { setPreferred2FAMethod } from "@services/UserInfo";
import MethodSelectionDialog from "@views/LoginPortal/SecondFactor/MethodSelectionDialog";

const OneTimePasswordMethod = lazy(() => import("@views/LoginPortal/SecondFactor/OneTimePasswordMethod"));
const PushNotificationMethod = lazy(() => import("@views/LoginPortal/SecondFactor/PushNotificationMethod"));
const WebAuthnMethod = lazy(() => import("@views/LoginPortal/SecondFactor/WebAuthnMethod"));

export interface Props {
    authenticationLevel: AuthenticationLevel;
    userInfo: UserInfo;
    configuration: Configuration;
    duoSelfEnrollment: boolean;

    onMethodChanged: () => void;
    onAuthenticationSuccess: (redirectURL: string | undefined) => void;
}

const SecondFactorForm = function (props: Props) {
    const styles = useStyles();
    const navigate = useNavigate();
    const [methodSelectionOpen, setMethodSelectionOpen] = useState(false);
    const [stateWebAuthnSupported, setStateWebAuthnSupported] = useState(false);
    const { createErrorNotification } = useNotifications();
    const { setLocalStorageMethod, localStorageMethodAvailable } = useLocalStorageMethodContext();
    const { t: translate } = useTranslation();

    useEffect(() => {
        setStateWebAuthnSupported(browserSupportsWebAuthn());
    }, [setStateWebAuthnSupported]);

    const handleMethodSelectionClick = () => {
        setMethodSelectionOpen(true);
    };

    const handleMethodSelected = async (method: SecondFactorMethod) => {
        if (localStorageMethodAvailable) {
            setLocalStorageMethod(method);
        } else {
            await handleMethodSelectedFallback(method);
        }

        setMethodSelectionOpen(false);
        props.onMethodChanged();
    };

    const handleMethodSelectedFallback = async (method: SecondFactorMethod) => {
        try {
            await setPreferred2FAMethod(method);
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("There was an issue updating preferred second factor method"));
        }
    };

    const handleLogoutClick = () => {
        navigate(SignOutRoute);
    };

    return (
        <LoginLayout
            id="second-factor-stage"
            title={`${translate("Hi")} ${props.userInfo.display_name}`}
            userInfo={props.userInfo}
        >
            {props.configuration.available_methods.size > 1 ? (
                <MethodSelectionDialog
                    open={methodSelectionOpen}
                    methods={props.configuration.available_methods}
                    webauthn={stateWebAuthnSupported}
                    onClose={() => setMethodSelectionOpen(false)}
                    onClick={handleMethodSelected}
                />
            ) : null}
            <Grid container>
                <Grid item xs={12}>
                    <Button color="secondary" onClick={handleLogoutClick} id="logout-button">
                        {translate("Logout")}
                    </Button>
                    {props.configuration.available_methods.size > 1 ? " | " : null}
                    {props.configuration.available_methods.size > 1 ? (
                        <Button color="secondary" onClick={handleMethodSelectionClick} id="methods-button">
                            {translate("Methods")}
                        </Button>
                    ) : null}
                </Grid>
                <Grid item xs={12} className={styles.methodContainer}>
                    <Routes>
                        <Route
                            path={SecondFactorTOTPSubRoute}
                            element={
                                <OneTimePasswordMethod
                                    id="one-time-password-method"
                                    authenticationLevel={props.authenticationLevel}
                                    // Whether the user has a TOTP secret registered already
                                    registered={props.userInfo.has_totp}
                                    onRegisterClick={() => {
                                        navigate(`${SettingsRoute}${SettingsTwoFactorAuthenticationSubRoute}`);
                                    }}
                                    onSignInError={(err) => createErrorNotification(err.message)}
                                    onSignInSuccess={props.onAuthenticationSuccess}
                                />
                            }
                        />
                        <Route
                            path={SecondFactorWebAuthnSubRoute}
                            element={
                                <WebAuthnMethod
                                    id="webauthn-method"
                                    authenticationLevel={props.authenticationLevel}
                                    // Whether the user has a WebAuthn device registered already
                                    registered={props.userInfo.has_webauthn}
                                    onRegisterClick={() => {
                                        navigate(`${SettingsRoute}${SettingsTwoFactorAuthenticationSubRoute}`);
                                    }}
                                    onSignInError={(err) => createErrorNotification(err.message)}
                                    onSignInSuccess={props.onAuthenticationSuccess}
                                />
                            }
                        />
                        <Route
                            path={SecondFactorPushSubRoute}
                            element={
                                <PushNotificationMethod
                                    id="push-notification-method"
                                    authenticationLevel={props.authenticationLevel}
                                    duoSelfEnrollment={props.duoSelfEnrollment}
                                    registered={props.userInfo.has_duo}
                                    onSelectionClick={props.onMethodChanged}
                                    onSignInError={(err) => createErrorNotification(err.message)}
                                    onSignInSuccess={props.onAuthenticationSuccess}
                                />
                            }
                        />
                    </Routes>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default SecondFactorForm;

const useStyles = makeStyles((theme: Theme) => ({
    methodContainer: {
        border: "1px solid #d6d6d6",
        borderRadius: "10px",
        padding: theme.spacing(4),
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));
