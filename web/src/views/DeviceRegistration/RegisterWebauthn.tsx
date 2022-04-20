import React, { useCallback, useEffect, useState } from "react";

import { Button, Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useLocation, useNavigate } from "react-router-dom";

import FingerTouchIcon from "@components/FingerTouchIcon";
import { useNotifications } from "@hooks/NotificationsContext";
import LoginLayout from "@layouts/LoginLayout";
import { AttestationResult } from "@models/Webauthn";
import { FirstFactorPath } from "@services/Api";
import { performAttestationCeremony } from "@services/Webauthn";
import { extractIdentityToken } from "@utils/IdentityToken";

const RegisterWebauthn = function () {
    const styles = useStyles();
    const navigate = useNavigate();
    const location = useLocation();
    const { createErrorNotification } = useNotifications();
    const [, setRegistrationInProgress] = useState(false);

    const processToken = extractIdentityToken(location.search);

    const handleBackClick = () => {
        navigate(FirstFactorPath);
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
        attestation();
    }, [attestation]);

    return (
        <LoginLayout title="Touch Security Key">
            <div className={styles.icon}>
                <FingerTouchIcon size={64} animated />
            </div>
            <Typography className={styles.instruction}>Touch the token on your security key</Typography>
            <Button color="primary" onClick={handleBackClick}>
                Retry
            </Button>
            <Button color="primary" onClick={handleBackClick}>
                Cancel
            </Button>
        </LoginLayout>
    );
};

export default RegisterWebauthn;

const useStyles = makeStyles((theme: Theme) => ({
    icon: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
    },
    instruction: {
        paddingBottom: theme.spacing(4),
    },
}));
