import React, { Fragment, MutableRefObject, useCallback, useEffect, useRef, useState } from "react";

import { Box, Button, Grid, Stack, Step, StepLabel, Stepper, Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { useLocation, useNavigate } from "react-router-dom";

import FixedTextField from "@components/FixedTextField";
import InformationIcon from "@components/InformationIcon";
import SuccessIcon from "@components/SuccessIcon";
import WebauthnTryIcon from "@components/WebauthnTryIcon";
import { SettingsRoute, SettingsTwoFactorAuthenticationSubRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import LoginLayout from "@layouts/LoginLayout";
import { AttestationPublicKeyCredential, AttestationResult, WebauthnTouchState } from "@models/Webauthn";
import {
    finishAttestationCeremony,
    getAttestationCreationOptions,
    getAttestationPublicKeyCredentialResult,
} from "@services/Webauthn";
import { extractIdentityToken } from "@utils/IdentityToken";

const steps = ["Confirm device", "Choose name"];

interface Props {}

const RegisterWebauthn = function (props: Props) {
    const [state, setState] = useState(WebauthnTouchState.WaitTouch);
    const styles = useStyles();
    const navigate = useNavigate();
    const location = useLocation();
    const { t: translate } = useTranslation();
    const { createErrorNotification } = useNotifications();

    const [activeStep, setActiveStep] = React.useState(0);
    const [credential, setCredential] = React.useState(null as null | AttestationPublicKeyCredential);
    const [creationOptions, setCreationOptions] = useState(null as null | PublicKeyCredentialCreationOptions);
    const [deviceName, setName] = useState("");
    const nameRef = useRef() as MutableRefObject<HTMLInputElement>;
    const [nameError, setNameError] = useState(false);

    const processToken = extractIdentityToken(location.search);

    const handleBackClick = () => {
        navigate(`${SettingsRoute}${SettingsTwoFactorAuthenticationSubRoute}`);
    };

    const finishAttestation = async () => {
        if (!credential) {
            return;
        }
        if (!deviceName.length) {
            setNameError(true);
            return;
        }
        const result = await finishAttestationCeremony(credential, deviceName);
        switch (result.status) {
            case AttestationResult.Success:
                setActiveStep(2);
                navigate(`${SettingsRoute}${SettingsTwoFactorAuthenticationSubRoute}`);
                break;
            case AttestationResult.Failure:
                createErrorNotification(result.message);
        }
    };

    const startAttestation = useCallback(async () => {
        try {
            setState(WebauthnTouchState.WaitTouch);
            setActiveStep(0);

            const startResult = await getAttestationPublicKeyCredentialResult(creationOptions);

            switch (startResult.result) {
                case AttestationResult.Success:
                    if (startResult.credential == null) {
                        throw new Error("Attestation request succeeded but credential is empty");
                    }
                    setCredential(startResult.credential);
                    setActiveStep(1);
                    return;
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
            setState(WebauthnTouchState.Failure);
        } catch (err) {
            console.error(err);
            createErrorNotification(
                "Failed to register your device. The identity verification process might have timed out.",
            );
        }
    }, [creationOptions, createErrorNotification]);

    useEffect(() => {
        if (creationOptions !== null) {
            startAttestation();
        }
    }, [creationOptions, startAttestation]);

    useEffect(() => {
        (async () => {
            const result = await getAttestationCreationOptions(processToken);
            if (result.status !== 200 || !result.options) {
                createErrorNotification(
                    "You must open the link from the same device and browser that initiated the registration process.",
                );
                return;
            }
            setCreationOptions(result.options);
        })();
    }, [processToken, setCreationOptions, createErrorNotification]);

    function renderStep(step: number) {
        switch (step) {
            case 0:
                return (
                    <Fragment>
                        <div className={styles.icon}>
                            <WebauthnTryIcon onRetryClick={startAttestation} webauthnTouchState={state} />
                        </div>
                        <Typography className={styles.instruction}>Touch the token on your security key</Typography>
                        <Grid container align="center" spacing={1}>
                            <Grid item xs={12}>
                                <Stack direction="row" spacing={1} justifyContent="center">
                                    <Button color="primary" onClick={handleBackClick}>
                                        Cancel
                                    </Button>
                                </Stack>
                            </Grid>
                        </Grid>
                    </Fragment>
                );
            case 1:
                return (
                    <div id="webauthn-registration-name">
                        <div className={styles.icon}>
                            <InformationIcon />
                        </div>
                        <Typography className={styles.instruction}>Enter a name for this key</Typography>
                        <Grid container spacing={1}>
                            <Grid item xs={12}>
                                <FixedTextField
                                    // TODO (PR: #806, Issue: #511) potentially refactor
                                    inputRef={nameRef}
                                    id="name-textfield"
                                    label={translate("Name")}
                                    variant="outlined"
                                    required
                                    value={deviceName}
                                    error={nameError}
                                    fullWidth
                                    disabled={false}
                                    onChange={(v) => setName(v.target.value.substring(0, 30))}
                                    onFocus={() => setNameError(false)}
                                    autoCapitalize="none"
                                    autoComplete="webauthn-name"
                                    onKeyPress={(ev) => {
                                        if (ev.key === "Enter") {
                                            if (!deviceName.length) {
                                                setNameError(true);
                                            } else {
                                                finishAttestation();
                                            }
                                            ev.preventDefault();
                                        }
                                    }}
                                />
                            </Grid>
                            <Grid item xs={12}>
                                <Stack direction="row" spacing={1} justifyContent="center">
                                    <Button color="primary" variant="outlined" onClick={startAttestation}>
                                        Back
                                    </Button>
                                    <Button color="primary" variant="contained" onClick={finishAttestation}>
                                        Finish
                                    </Button>
                                </Stack>
                            </Grid>
                        </Grid>
                    </div>
                );
            case 2:
                return (
                    <div id="webauthn-registration-success">
                        <div className={styles.iconContainer}>
                            <SuccessIcon />
                        </div>
                        <Typography>{translate("Registration success")}</Typography>
                    </div>
                );
        }
    }

    return (
        <LoginLayout title="Register Security Key">
            <Grid container>
                <Grid item xs={12} className={styles.methodContainer}>
                    <Box sx={{ width: "100%" }}>
                        <Stepper activeStep={activeStep}>
                            {steps.map((label, index) => {
                                const stepProps: { completed?: boolean } = {};
                                const labelProps: {
                                    optional?: React.ReactNode;
                                } = {};
                                return (
                                    <Step key={label} {...stepProps}>
                                        <StepLabel {...labelProps}>{label}</StepLabel>
                                    </Step>
                                );
                            })}
                        </Stepper>
                        {renderStep(activeStep)}
                    </Box>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default RegisterWebauthn;

const useStyles = makeStyles((theme: Theme) => ({
    icon: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
    },
    iconContainer: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%",
    },
    instruction: {
        paddingBottom: theme.spacing(4),
    },
    methodContainer: {
        border: "1px solid #d6d6d6",
        borderRadius: "10px",
        padding: theme.spacing(4),
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));
