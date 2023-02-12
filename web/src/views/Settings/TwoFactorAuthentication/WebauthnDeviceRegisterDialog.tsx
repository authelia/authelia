import React, { Fragment, MutableRefObject, useCallback, useEffect, useRef, useState } from "react";

import {
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Grid,
    Stack,
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
import {
    AttestationResult,
    AttestationResultFailureString,
    RegistrationResult,
    WebauthnTouchState,
} from "@models/Webauthn";
import { finishRegistration, getAttestationCreationOptions, startWebauthnRegistration } from "@services/Webauthn";

const steps = ["Confirm device", "Choose name"];

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
    const [result, setResult] = useState<RegistrationResult | null>(null);
    const [options, setOptions] = useState<PublicKeyCredentialCreationOptionsJSON | null>(null);
    const [timeout, setTimeout] = useState<number | null>(null);
    const [deviceName, setName] = useState("");

    const nameRef = useRef() as MutableRefObject<HTMLInputElement>;
    const [nameError, setNameError] = useState(false);

    const resetStates = () => {
        setState(WebauthnTouchState.WaitTouch);
        setActiveStep(0);
        setResult(null);
        setOptions(null);
        setTimeout(null);
        setName("");
    };

    const handleClose = useCallback(() => {
        resetStates();

        props.setCancelled();
    }, [props]);

    const finishAttestation = async () => {
        if (!result || !result.response) {
            return;
        }

        if (!deviceName.length) {
            setNameError(true);
            return;
        }

        const res = await finishRegistration(result.response, deviceName);
        switch (res.status) {
            case AttestationResult.Success:
                handleClose();
                break;
            case AttestationResult.Failure:
                createErrorNotification(res.message);
        }
    };

    const startRegistration = useCallback(async () => {
        if (options === null) {
            return;
        }

        setTimeout(options.timeout ? options.timeout : null);

        try {
            setState(WebauthnTouchState.WaitTouch);
            setActiveStep(0);

            const res = await startWebauthnRegistration(options);

            setTimeout(null);

            if (res.result === AttestationResult.Success) {
                if (res.response == null) {
                    throw new Error("Attestation request succeeded but credential is empty");
                }

                setResult(res);
                setActiveStep(1);

                return;
            }

            createErrorNotification(AttestationResultFailureString(res.result));
            setState(WebauthnTouchState.Failure);
        } catch (err) {
            console.error(err);
            createErrorNotification(
                "Failed to register your device. The identity verification process might have timed out.",
            );
        }
    }, [options, createErrorNotification]);

    useEffect(() => {
        if (state !== WebauthnTouchState.Failure || activeStep !== 0 || !props.open) {
            return;
        }

        handleClose();
    }, [props, state, activeStep, handleClose]);

    useEffect(() => {
        (async () => {
            if (options === null || !props.open || activeStep !== 0) {
                return;
            }

            await startRegistration();
        })();
    }, [options, props.open, activeStep, startRegistration]);

    useEffect(() => {
        (async () => {
            if (!props.open || activeStep !== 0) {
                return;
            }

            const res = await getAttestationCreationOptions();
            if (res.status !== 200 || !res.options) {
                createErrorNotification(
                    "You must open the link from the same device and browser that initiated the registration process.",
                );
                return;
            }
            setOptions(res.options);
        })();
    }, [setOptions, createErrorNotification, props.open, activeStep]);

    function renderStep(step: number) {
        switch (step) {
            case 0:
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
            case 1:
                return (
                    <Box id="webauthn-registration-name">
                        <Box className={styles.icon}>
                            <InformationIcon />
                        </Box>
                        <Typography className={styles.instruction}>{translate("Enter a name for this key")}</Typography>
                        <Grid container spacing={1}>
                            <Grid item xs={12}>
                                <TextField
                                    inputRef={nameRef}
                                    id="name-textfield"
                                    label={translate("Name")}
                                    variant="outlined"
                                    required
                                    value={deviceName}
                                    error={nameError}
                                    disabled={false}
                                    onChange={(v) => setName(v.target.value.substring(0, 30))}
                                    onFocus={() => setNameError(false)}
                                    autoCapitalize="none"
                                    autoComplete="webauthn-name"
                                    onKeyDown={(ev) => {
                                        if (ev.key === "Enter") {
                                            if (!deviceName.length) {
                                                setNameError(true);
                                            } else {
                                                (async () => {
                                                    await finishAttestation();
                                                })();
                                            }
                                            ev.preventDefault();
                                        }
                                    }}
                                />
                            </Grid>
                            <Grid item xs={12}>
                                <Stack direction="row" spacing={1} justifyContent="center" paddingTop={1}>
                                    <Button color="primary" variant="contained" onClick={finishAttestation}>
                                        {translate("Finish")}
                                    </Button>
                                </Stack>
                            </Grid>
                        </Grid>
                    </Box>
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
            <DialogTitle>{translate("Register Webauthn Credential (Security Key)")}</DialogTitle>
            <DialogContent>
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
                <Button onClick={handleClose} disabled={activeStep === 0 && state !== WebauthnTouchState.Failure}>
                    {translate("Cancel")}
                </Button>
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
