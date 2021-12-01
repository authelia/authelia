import React, { useState, useEffect } from "react";

import { Grid, makeStyles, Button } from "@material-ui/core";
import { Route, Routes, useNavigate } from "react-router-dom";
import u2fApi from "u2f-api";

import {
    LogoutRoute as SignOutRoute,
    SecondFactorPushSubRoute,
    SecondFactorTOTPSubRoute,
    SecondFactorU2FSubRoute,
} from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import LoginLayout from "@layouts/LoginLayout";
import { Configuration } from "@models/Configuration";
import { SecondFactorMethod } from "@models/Methods";
import { UserInfo } from "@models/UserInfo";
import { initiateTOTPRegistrationProcess, initiateU2FRegistrationProcess } from "@services/RegisterDevice";
import { AuthenticationLevel } from "@services/State";
import { setPreferred2FAMethod } from "@services/UserInfo";
import MethodSelectionDialog from "@views/LoginPortal/SecondFactor/MethodSelectionDialog";
import OneTimePasswordMethod from "@views/LoginPortal/SecondFactor/OneTimePasswordMethod";
import PushNotificationMethod from "@views/LoginPortal/SecondFactor/PushNotificationMethod";
import SecurityKeyMethod from "@views/LoginPortal/SecondFactor/SecurityKeyMethod";

const EMAIL_SENT_NOTIFICATION = "An email has been sent to your address to complete the process.";

export interface Props {
    authenticationLevel: AuthenticationLevel;
    userInfo: UserInfo;
    configuration: Configuration;
    duoSelfEnrollment: boolean;

    onMethodChanged: () => void;
    onAuthenticationSuccess: (redirectURL: string | undefined) => void;
}

const SecondFactorForm = function (props: Props) {
    const style = useStyles();
    const navigate = useNavigate();
    const [methodSelectionOpen, setMethodSelectionOpen] = useState(false);
    const { createInfoNotification, createErrorNotification } = useNotifications();
    const [registrationInProgress, setRegistrationInProgress] = useState(false);
    const [u2fSupported, setU2fSupported] = useState(false);

    // Check that U2F is supported.
    useEffect(() => {
        u2fApi.ensureSupport().then(
            () => setU2fSupported(true),
            () => console.error("U2F not supported"),
        );
    }, [setU2fSupported]);

    const initiateRegistration = (initiateRegistrationFunc: () => Promise<void>) => {
        return async () => {
            if (registrationInProgress) {
                return;
            }
            setRegistrationInProgress(true);
            try {
                await initiateRegistrationFunc();
                createInfoNotification(EMAIL_SENT_NOTIFICATION);
            } catch (err) {
                console.error(err);
                createErrorNotification("There was a problem initiating the registration process");
            }
            setRegistrationInProgress(false);
        };
    };

    const handleMethodSelectionClick = () => {
        setMethodSelectionOpen(true);
    };

    const handleMethodSelected = async (method: SecondFactorMethod) => {
        try {
            await setPreferred2FAMethod(method);
            setMethodSelectionOpen(false);
            props.onMethodChanged();
        } catch (err) {
            console.error(err);
            createErrorNotification("There was an issue updating preferred second factor method");
        }
    };

    const handleLogoutClick = () => {
        navigate(SignOutRoute);
    };

    return (
        <LoginLayout id="second-factor-stage" title={`Hi ${props.userInfo.display_name}`} showBrand>
            <MethodSelectionDialog
                open={methodSelectionOpen}
                methods={props.configuration.available_methods}
                u2fSupported={u2fSupported}
                onClose={() => setMethodSelectionOpen(false)}
                onClick={handleMethodSelected}
            />
            <Grid container>
                <Grid item xs={12}>
                    <Button color="secondary" onClick={handleLogoutClick} id="logout-button">
                        Logout
                    </Button>
                    {" | "}
                    <Button color="secondary" onClick={handleMethodSelectionClick} id="methods-button">
                        Methods
                    </Button>
                </Grid>
                <Grid item xs={12} className={style.methodContainer}>
                    <Routes>
                        <Route
                            path={SecondFactorTOTPSubRoute}
                            element={
                                <OneTimePasswordMethod
                                    id="one-time-password-method"
                                    authenticationLevel={props.authenticationLevel}
                                    // Whether the user has a TOTP secret registered already
                                    registered={props.userInfo.has_totp}
                                    onRegisterClick={initiateRegistration(initiateTOTPRegistrationProcess)}
                                    onSignInError={(err) => createErrorNotification(err.message)}
                                    onSignInSuccess={props.onAuthenticationSuccess}
                                />
                            }
                        />
                        <Route
                            path={SecondFactorU2FSubRoute}
                            element={
                                <SecurityKeyMethod
                                    id="security-key-method"
                                    authenticationLevel={props.authenticationLevel}
                                    // Whether the user has a U2F device registered already
                                    registered={props.userInfo.has_u2f}
                                    onRegisterClick={initiateRegistration(initiateU2FRegistrationProcess)}
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

const useStyles = makeStyles((theme) => ({
    methodContainer: {
        border: "1px solid #d6d6d6",
        borderRadius: "10px",
        padding: theme.spacing(4),
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));
