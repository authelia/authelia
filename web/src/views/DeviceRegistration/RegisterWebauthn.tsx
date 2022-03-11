import React, { useCallback, useEffect, useState } from "react";

import { Button, makeStyles, Typography } from "@material-ui/core";
import { useLocation, useNavigate } from "react-router-dom";
import UAParser from "ua-parser-js";

import FingerTouchIcon from "@components/FingerTouchIcon";
import { useNotifications } from "@hooks/NotificationsContext";
import LoginLayout from "@layouts/LoginLayout";
import { AttestationResult } from "@models/Webauthn";
import { FirstFactorPath } from "@services/Api";
import { performAttestationCeremony } from "@services/Webauthn";
import { extractIdentityToken } from "@utils/IdentityToken";

export enum State {
    InProgress = 1,
    UserGestureRequired = 2,
}

const RegisterWebauthn = function () {
    const uaParser = new UAParser();

    const uaResult = uaParser.getResult();

    const apple =
        uaResult.device.vendor === "Apple" ||
        uaResult.os.name === "iOS" ||
        uaResult.os.name === "Mac OS" ||
        uaResult.browser.name === "Safari";

    console.log("device.vendor: ", uaResult.device.vendor);
    console.log("device.model: ", uaResult.device.model);
    console.log("device.type: ", uaResult.device.type);
    console.log("cpu.architecture: ", uaResult.cpu.architecture);
    console.log("os.name: ", uaResult.os.name);
    console.log("os.version: ", uaResult.os.version);
    console.log("browser.name: ", uaResult.browser.version);
    console.log("browser.version: ", uaResult.browser.version);
    console.log("engine.name: ", uaResult.engine.name);
    console.log("engine.version: ", uaResult.engine.version);
    console.log("apple: ", apple);

    const [state, setState] = useState(apple ? State.UserGestureRequired : State.InProgress);

    const style = useStyles();
    const navigate = useNavigate();
    const location = useLocation();
    const { createErrorNotification } = useNotifications();
    const [, setRegistrationInProgress] = useState(false);

    const processToken = extractIdentityToken(location.search);

    const handleBackClick = () => {
        navigate(FirstFactorPath);
    };

    const handleStartClick = () => {
        setState(State.InProgress);

        attestation();
    };

    const attestation = useCallback(async () => {
        if (!processToken) {
            return;
        }
        try {
            setRegistrationInProgress(true);

            const result = await performAttestationCeremony(processToken);

            setRegistrationInProgress(false);

            switch (result) {
                case AttestationResult.Success:
                    navigate(FirstFactorPath);
                    break;
                case AttestationResult.FailureToken:
                    createErrorNotification(
                        "You must open the link from the same device and browser that initiated the registration process.",
                    );
                    break;
                case AttestationResult.FailureSupport:
                    createErrorNotification("Your browser does not appear to support the configuration.");
                    break;
                case AttestationResult.FailureSyntax:
                    createErrorNotification(
                        "The attestation challenge was rejected as malformed or incompatible by your browser.",
                    );
                    break;
                case AttestationResult.FailureWebauthnNotSupported:
                    createErrorNotification("Your browser does not support the WebAuthN protocol.");
                    break;
                case AttestationResult.FailureUserConsent:
                    createErrorNotification("You cancelled the attestation request.");
                    break;
                case AttestationResult.FailureUserVerificationOrResidentKey:
                    createErrorNotification(
                        "Your device does not support user verification or resident keys but this was required.",
                    );
                    break;
                case AttestationResult.FailureExcluded:
                    createErrorNotification("You have registered this device already.");
                    break;
                case AttestationResult.FailureUnknown:
                    createErrorNotification("An unknown error occurred.");
                    break;
            }
        } catch (err) {
            console.error(err);
            createErrorNotification(
                "Failed to register your device. The identity verification process might have timed out.",
            );
        }
    }, [processToken, createErrorNotification, navigate]);

    useEffect(() => {
        if (state === State.InProgress) {
            attestation();
        }
    }, [attestation, state]);

    return (
        <LoginLayout title="Touch Security Key">
            <div className={style.icon}>
                <FingerTouchIcon size={64} animated />
            </div>
            {state === State.InProgress ? (
                <div>
                    <Typography className={style.instruction}>Touch the token on your security key</Typography>

                    <Button color="primary" onClick={handleBackClick}>
                        Retry
                    </Button>
                    <Button color="primary" onClick={handleBackClick}>
                        Cancel
                    </Button>
                </div>
            ) : (
                <div>
                    <Typography className={style.instruction}>Click start to begin</Typography>

                    <Button color="primary" onClick={handleStartClick}>
                        Start
                    </Button>
                    <Button color="primary" onClick={handleBackClick}>
                        Cancel
                    </Button>
                </div>
            )}
        </LoginLayout>
    );
};

export default RegisterWebauthn;

const useStyles = makeStyles((theme) => ({
    icon: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
    },
    instruction: {
        paddingBottom: theme.spacing(4),
    },
}));
