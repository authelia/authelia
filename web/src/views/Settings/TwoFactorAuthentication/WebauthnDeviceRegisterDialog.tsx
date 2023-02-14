import React, { Fragment, MutableRefObject, useCallback, useEffect, useRef, useState } from "react";

import {
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Grid,
    Step,
    StepLabel,
    Stepper,
    TextField,
    Theme,
    Typography,
} from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { PublicKeyCredentialCreationOptionsJSON } from "@simplewebauthn/typescript-types";
import { useTranslation } from "react-i18next";

import InformationIcon from "@components/InformationIcon";
import WebauthnRegisterIcon from "@components/WebauthnRegisterIcon";
import { useNotifications } from "@hooks/NotificationsContext";
import { AttestationResult, AttestationResultFailureString, WebauthnTouchState } from "@models/Webauthn";
import { finishRegistration, getAttestationCreationOptions, startWebauthnRegistration } from "@services/Webauthn";

const steps = ["Display Name", "Verification"];

interface Props {
    open: boolean;
    onClose: () => void;
    setCancelled: () => void;
}

const WebauthnDeviceRegisterDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const styles = useStyles();
    const { createErrorNotification } = useNotifications();

    const [state, setState] = useState(WebauthnTouchState.WaitTouch);
    const [activeStep, setActiveStep] = useState(0);
    const [options, setOptions] = useState<PublicKeyCredentialCreationOptionsJSON | null>(null);
    const [timeout, setTimeout] = useState<number | null>(null);
    const [credentialDisplayName, setCredentialDisplayName] = useState("");
    const [errorDisplayName, setErrorDisplayName] = useState(false);

    const nameRef = useRef() as MutableRefObject<HTMLInputElement>;

    const resetStates = () => {
        setState(WebauthnTouchState.WaitTouch);
        setOptions(null);
        setActiveStep(0);
        setTimeout(null);
        setCredentialDisplayName("");
        setErrorDisplayName(false);
    };

    const handleClose = useCallback(() => {
        resetStates();

        props.setCancelled();
    }, [props]);

    const performCredentialCreation = useCallback(async () => {
        if (options === null) {
            return;
        }

        setTimeout(options.timeout ? options.timeout : null);
        setActiveStep(1);

        try {
            setState(WebauthnTouchState.WaitTouch);

            const resultCredentialCreation = await startWebauthnRegistration(options);

            setTimeout(null);

            if (resultCredentialCreation.result === AttestationResult.Success) {
                if (resultCredentialCreation.response == null) {
                    throw new Error("Credential Creation Request succeeded but Registration Response is empty.");
                }

                const response = await finishRegistration(resultCredentialCreation.response);

                switch (response.status) {
                    case AttestationResult.Success:
                        handleClose();
                        break;
                    case AttestationResult.Failure:
                        createErrorNotification(response.message);
                        break;
                }

                return;
            }

            createErrorNotification(AttestationResultFailureString(resultCredentialCreation.result));
            setState(WebauthnTouchState.Failure);
        } catch (err) {
            console.error(err);
            createErrorNotification(
                "Failed to register your device. The identity verification process might have timed out.",
            );
        }
    }, [options, createErrorNotification, handleClose]);

    useEffect(() => {
        if (state !== WebauthnTouchState.Failure || activeStep !== 0 || !props.open) {
            return;
        }

        handleClose();
    }, [props, state, activeStep, handleClose]);

    useEffect(() => {
        (async function () {
            if (!props.open || activeStep !== 0 || options === null) {
                return;
            }

            await performCredentialCreation();
        })();
    }, [props.open, activeStep, options, performCredentialCreation]);

    const handleNext = useCallback(async () => {
        if (credentialDisplayName.length === 0 || credentialDisplayName.length > 64) {
            setErrorDisplayName(true);
            createErrorNotification(
                translate("The Display Name must be more than 1 character and less than 64 characters."),
            );

            return;
        }

        const res = await getAttestationCreationOptions(credentialDisplayName);

        switch (res.status) {
            case 200:
                if (res.options) {
                    setOptions(res.options);
                } else {
                    throw new Error(
                        "Credential Creation Options Request succeeded but Credential Creation Options is empty.",
                    );
                }

                break;
            case 409:
                setErrorDisplayName(true);
                createErrorNotification(translate("A Webauthn Credential with that Display Name already exists."));

                break;
            default:
                createErrorNotification(
                    translate("Error occurred obtaining the Webauthn Credential creation options."),
                );
        }
    }, [createErrorNotification, credentialDisplayName, performCredentialCreation, translate]);

    const handleCredentialDisplayName = useCallback(
        (displayname: string) => {
            setCredentialDisplayName(displayname);

            if (errorDisplayName) {
                setErrorDisplayName(false);
            }
        },
        [errorDisplayName],
    );

    function renderStep(step: number) {
        switch (step) {
            case 0:
                return (
                    <Box>
                        <Box className={styles.icon}>
                            <InformationIcon />
                        </Box>
                        <Typography className={styles.instruction}>
                            {translate("Enter a display name for this credential")}
                        </Typography>
                        <Grid container spacing={1}>
                            <Grid item xs={12}>
                                <TextField
                                    inputRef={nameRef}
                                    id="name-textfield"
                                    label={translate("Display Name")}
                                    variant="outlined"
                                    required
                                    value={credentialDisplayName}
                                    error={errorDisplayName}
                                    disabled={false}
                                    onChange={(v) => handleCredentialDisplayName(v.target.value)}
                                    autoCapitalize="none"
                                    onKeyDown={(ev) => {
                                        if (ev.key === "Enter") {
                                            (async () => {
                                                await handleNext();
                                            })();

                                            ev.preventDefault();
                                        }
                                    }}
                                />
                            </Grid>
                        </Grid>
                    </Box>
                );
            case 1:
                return (
                    <Fragment>
                        <Box className={styles.icon}>
                            {timeout !== null ? <WebauthnRegisterIcon timeout={timeout} /> : null}
                        </Box>
                        <Typography className={styles.instruction}>
                            {translate("Touch the token on your security key")}
                        </Typography>
                    </Fragment>
                );
        }
    }

    const handleOnClose = () => {
        if (activeStep === 0 || !props.open) {
            return;
        }

        handleClose();
    };

    return (
        <Dialog open={props.open} onClose={handleOnClose} maxWidth={"xs"} fullWidth={true}>
            <DialogTitle>{translate("Register Webauthn Credential")}</DialogTitle>
            <DialogContent>
                <DialogContentText sx={{ mb: 3 }}>
                    {translate(
                        "This page allows registration of a new Security Key backed by modern Webauthn Credential technology.",
                    )}
                </DialogContentText>
                <Grid container spacing={0} alignItems={"center"} justifyContent={"center"} textAlign={"center"}>
                    <Grid item xs={12}>
                        <Stepper activeStep={activeStep}>
                            {steps.map((label, index) => {
                                const stepProps: { completed?: boolean } = {};
                                const labelProps: {
                                    optional?: React.ReactNode;
                                } = {};
                                return (
                                    <Step key={label} {...stepProps}>
                                        <StepLabel {...labelProps}>{translate(label)}</StepLabel>
                                    </Step>
                                );
                            })}
                        </Stepper>
                    </Grid>
                    <Grid item xs={12}>
                        {renderStep(activeStep)}
                    </Grid>
                </Grid>
            </DialogContent>
            <DialogActions>
                <Button
                    color={activeStep === 1 && state !== WebauthnTouchState.Failure ? "primary" : "error"}
                    disabled={activeStep === 1 && state !== WebauthnTouchState.Failure}
                    onClick={handleClose}
                >
                    {translate("Cancel")}
                </Button>
                {activeStep === 0 ? (
                    <Button
                        color={credentialDisplayName.length !== 0 ? "success" : "primary"}
                        disabled={activeStep !== 0}
                        onClick={async () => {
                            await handleNext();
                        }}
                    >
                        {translate("Next")}
                    </Button>
                ) : null}
            </DialogActions>
        </Dialog>
    );
};

export default WebauthnDeviceRegisterDialog;

const useStyles = makeStyles((theme: Theme) => ({
    icon: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
    },
    instruction: {
        paddingBottom: theme.spacing(4),
    },
}));
