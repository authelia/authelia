import React, { Fragment, useCallback, useEffect, useRef, useState } from "react";

import {
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Step,
    StepLabel,
    Stepper,
    TextField,
    Theme,
    Typography,
} from "@mui/material";
import Grid from "@mui/material/Grid";
import { PublicKeyCredentialCreationOptionsJSON } from "@simplewebauthn/browser";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import InformationIcon from "@components/InformationIcon";
import WebAuthnRegisterIcon from "@components/WebAuthnRegisterIcon";
import { useNotifications } from "@hooks/NotificationsContext";
import { AttestationResult, AttestationResultFailureString, WebAuthnTouchState } from "@models/WebAuthn";
import {
    finishWebAuthnRegistration,
    getWebAuthnRegistrationOptions,
    startWebAuthnRegistration,
} from "@services/WebAuthn";

const steps = ["Description", "Verification"];

interface Props {
    open: boolean;
    setClosed: () => void;
}

const WebAuthnCredentialRegisterDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { classes } = useStyles();

    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const [state, setState] = useState(WebAuthnTouchState.WaitTouch);
    const [activeStep, setActiveStep] = useState(0);
    const [options, setOptions] = useState<PublicKeyCredentialCreationOptionsJSON | null>(null);
    const [timeout, setTimeout] = useState<number | null>(null);
    const [description, setDescription] = useState("");
    const [errorDescription, setErrorDescription] = useState(false);

    const nameRef = useRef<HTMLInputElement | null>(null);

    const resetStates = () => {
        setState(WebAuthnTouchState.WaitTouch);
        setOptions(null);
        setActiveStep(0);
        setTimeout(null);
        setDescription("");
        setErrorDescription(false);
    };

    const handleClose = useCallback(() => {
        resetStates();

        props.setClosed();
    }, [props]);

    const performCredentialCreation = useCallback(async () => {
        if (!props.open || options === null) {
            return;
        }

        setTimeout(options.timeout ? options.timeout : null);
        setActiveStep(1);

        try {
            setState(WebAuthnTouchState.WaitTouch);

            const result = await startWebAuthnRegistration(options);

            setTimeout(null);

            if (result.result === AttestationResult.Success) {
                if (result.response == null) {
                    throw new Error("Credential Creation Request succeeded but Registration Response is empty.");
                }

                const response = await finishWebAuthnRegistration(result.response);

                switch (response.status) {
                    case AttestationResult.Success:
                        createSuccessNotification(
                            translate("Successfully {{action}} the {{item}}", {
                                action: translate("added"),
                                item: translate("WebAuthn Credential"),
                            }),
                        );
                        break;
                    case AttestationResult.Failure:
                        createErrorNotification(response.message);
                        break;
                }

                return;
            } else {
                createErrorNotification(translate(AttestationResultFailureString(result.result)));
                setState(WebAuthnTouchState.Failure);
            }
        } catch (err) {
            console.error(err);
            createErrorNotification(
                translate("Failed to register your credential, the identity verification process might have timed out"),
            );
        } finally {
            handleClose();
        }
    }, [props.open, options, createSuccessNotification, translate, createErrorNotification, handleClose]);

    useEffect(() => {
        if (!props.open || state !== WebAuthnTouchState.Failure || activeStep !== 0) {
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

    const handleNext = useCallback(() => {
        if (!props.open) {
            return;
        }

        (async function () {
            if (description.length === 0 || description.length > 64) {
                setErrorDescription(true);
                createErrorNotification(
                    translate("The Description must be more than 1 character and less than 64 characters"),
                );

                return;
            }

            const res = await getWebAuthnRegistrationOptions(description);

            switch (res.status) {
                case 200:
                    if (res.options) {
                        setOptions(res.options);
                    } else {
                        createErrorNotification(
                            translate(
                                "Credential Creation Options Request succeeded but Credential Creation Options is empty",
                            ),
                        );
                    }

                    break;
                case 409:
                    setErrorDescription(true);
                    createErrorNotification(translate("A WebAuthn Credential with that Description already exists"));

                    break;
                default:
                    createErrorNotification(
                        translate("Error occurred obtaining the WebAuthn Credential creation options"),
                    );
            }

            await performCredentialCreation();
        })();
    }, [createErrorNotification, description, performCredentialCreation, props.open, translate]);

    const handleCredentialDescription = useCallback(
        (description: string) => {
            setDescription(description);

            if (errorDescription) {
                setErrorDescription(false);
            }
        },
        [errorDescription],
    );

    function renderStep(step: number) {
        switch (step) {
            case 0:
                return (
                    <Box>
                        <Box className={classes.icon}>
                            <InformationIcon />
                        </Box>
                        <Typography className={classes.instruction}>
                            {translate("Enter a description for this WebAuthn Credential")}
                        </Typography>
                        <Grid container spacing={1}>
                            <Grid size={{ xs: 12 }}>
                                <TextField
                                    inputRef={nameRef}
                                    id="webauthn-credential-description"
                                    label={translate("Description")}
                                    variant="outlined"
                                    required
                                    value={description}
                                    error={errorDescription}
                                    disabled={false}
                                    onChange={(v) => handleCredentialDescription(v.target.value)}
                                    autoCapitalize="none"
                                    onKeyDown={(ev) => {
                                        if (ev.key === "Enter") {
                                            (async () => {
                                                handleNext();
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
                        <Box className={classes.icon}>
                            {timeout !== null ? <WebAuthnRegisterIcon timeout={timeout} /> : null}
                        </Box>
                        <Typography className={classes.instruction}>
                            {translate("Touch the token on your security key")}
                        </Typography>
                    </Fragment>
                );
        }
    }

    const handleOnClose = () => {
        if (!props.open || activeStep === 1) {
            return;
        }

        handleClose();
    };

    return (
        <Dialog open={props.open} onClose={handleOnClose} maxWidth={"xs"} fullWidth>
            <DialogTitle>{translate("Register {{item}}", { item: translate("WebAuthn Credential") })}</DialogTitle>
            <DialogContent>
                <DialogContentText sx={{ mb: 3 }}>
                    {translate("This dialog handles registration of a {{item}}", {
                        item: translate("WebAuthn Credential"),
                    })}
                </DialogContentText>
                <Grid container spacing={0} alignItems={"center"} justifyContent={"center"} textAlign={"center"}>
                    <Grid size={{ xs: 12 }}>
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
                    <Grid size={{ xs: 12 }}>{renderStep(activeStep)}</Grid>
                </Grid>
            </DialogContent>
            <DialogActions>
                <Button
                    id={"dialog-cancel"}
                    color={activeStep === 1 && state !== WebAuthnTouchState.Failure ? "primary" : "error"}
                    disabled={activeStep === 1 && state !== WebAuthnTouchState.Failure}
                    onClick={handleClose}
                    data-1p-ignore
                >
                    {translate("Cancel")}
                </Button>
                {activeStep === 0 ? (
                    <Button
                        id={"dialog-next"}
                        color={description.length !== 0 ? "success" : "primary"}
                        disabled={activeStep !== 0}
                        onClick={async () => {
                            handleNext();
                        }}
                        data-1p-ignore
                    >
                        {translate("Next")}
                    </Button>
                ) : null}
            </DialogActions>
        </Dialog>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    icon: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
    },
    instruction: {
        paddingBottom: theme.spacing(4),
    },
}));

export default WebAuthnCredentialRegisterDialog;
