import React, { lazy, useEffect, useState } from "react";

import { Box, Button, Theme } from "@mui/material";
import Grid from "@mui/material/Grid2";
import makeStyles from "@mui/styles/makeStyles";
import { browserSupportsWebAuthn } from "@simplewebauthn/browser";
import { useTranslation } from "react-i18next";
import { Route, Routes, useNavigate } from "react-router-dom";

import {
    SecondFactorPasswordSubRoute,
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
const PasswordMethod = lazy(() => import("@views/LoginPortal/SecondFactor/PasswordMethod"));

export interface Props {
    authenticationLevel: AuthenticationLevel;
    factorKnowledge: boolean;
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

    const showMethods = props.factorKnowledge && props.configuration.available_methods.size > 1;

    return (
        <LoginLayout
            id={"second-factor-stage"}
            title={`${translate("Hi")} ${props.userInfo.display_name}`}
            userInfo={props.userInfo}
        >
            {showMethods ? (
                <MethodSelectionDialog
                    open={methodSelectionOpen}
                    methods={props.configuration.available_methods}
                    webauthn={stateWebAuthnSupported}
                    onClose={() => setMethodSelectionOpen(false)}
                    onClick={handleMethodSelected}
                />
            ) : null}
            <Grid container direction={"column"} justifyContent={"center"} alignItems={"center"}>
                <Grid size={{ xs: 12 }}>
                    <Button id={"logout-button"} color={"secondary"} onClick={handleLogoutClick}>
                        {translate("Logout")}
                    </Button>
                    {showMethods ? " | " : null}
                    {showMethods ? (
                        <Button id={"methods-button"} color="secondary" onClick={handleMethodSelectionClick}>
                            {translate("Methods")}
                        </Button>
                    ) : null}
                </Grid>
                <Box className={styles.methodContainer}>
                    <Routes>
                        <Route
                            path={SecondFactorPasswordSubRoute}
                            element={
                                <PasswordMethod
                                    id="password-method"
                                    authenticationLevel={props.authenticationLevel}
                                    userInfo={props.userInfo}
                                    onAuthenticationSuccess={props.onAuthenticationSuccess}
                                />
                            }
                        />
                        <Route
                            path={SecondFactorTOTPSubRoute}
                            element={
                                <OneTimePasswordMethod
                                    id={"one-time-password-method"}
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
                                    id={"webauthn-method"}
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
                                    id={"push-notification-method"}
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
                </Box>
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
        minWidth: "300px",
    },
}));
