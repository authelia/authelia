import React, { Fragment, useCallback, useEffect, useState } from "react";

import { faTimesCircle } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
    Box,
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    FormControl,
    FormControlLabel,
    FormLabel,
    Link,
    Radio,
    RadioGroup,
    Step,
    StepLabel,
    Stepper,
    Switch,
    TextField,
    Theme,
    Typography,
} from "@mui/material";
import { red } from "@mui/material/colors";
import Grid from "@mui/material/Grid";
import { QRCodeSVG } from "qrcode.react";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import AppStoreBadges from "@components/AppStoreBadges";
import CopyButton from "@components/CopyButton";
import SuccessIcon from "@components/SuccessIcon";
import { GoogleAuthenticator } from "@constants/constants";
import { useNotifications } from "@hooks/NotificationsContext";
import { toAlgorithmString } from "@models/TOTPConfiguration";
import { completeTOTPRegister, stopTOTPRegister } from "@services/OneTimePassword";
import { getTOTPSecret } from "@services/RegisterDevice";
import { getTOTPOptions } from "@services/UserInfoTOTPConfiguration";
import OTPDial, { State } from "@views/LoginPortal/SecondFactor/OTPDial";

const steps = ["Start", "Register", "Confirm"];

interface Props {
    open: boolean;
    setClosed: () => void;
}

interface Options {
    algorithm: string;
    length: number;
    period: number;
}

interface AvailableOptions {
    algorithms: string[];
    lengths: number[];
    periods: number[];
}

const OneTimePasswordRegisterDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const { classes, cx } = useStyles();
    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const [selected, setSelected] = useState<Options>({ algorithm: "", length: 6, period: 30 });
    const [defaults, setDefaults] = useState<Options | null>(null);
    const [available, setAvailable] = useState<AvailableOptions>({
        algorithms: [],
        lengths: [],
        periods: [],
    });

    const [activeStep, setActiveStep] = useState(0);

    const [secretURL, setSecretURL] = useState<string | null>(null);
    const [secretValue, setSecretValue] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [showAdvanced, setShowAdvanced] = useState(false);
    const [hasErrored, setHasErrored] = useState(false);
    const [dialValue, setDialValue] = useState("");
    const [dialState, setDialState] = useState(State.Idle);
    const [showQRCode, setShowQRCode] = useState(true);
    const [success, setSuccess] = useState(false);

    const resetStates = useCallback(() => {
        if (defaults) {
            setSelected(defaults);
        }

        setSecretURL(null);
        setSecretValue(null);
        setIsLoading(false);
        setShowAdvanced(false);
        setHasErrored(false);
        setActiveStep(0);
        setDialValue("");
        setDialState(State.Idle);
        setShowQRCode(true);
        setSuccess(false);
    }, [defaults]);

    const handleClose = useCallback(() => {
        (async () => {
            props.setClosed();

            if (secretURL !== "") {
                try {
                    await stopTOTPRegister();
                } catch (err) {
                    console.error(err);
                }
            }

            resetStates();
        })();
    }, [props, secretURL, resetStates]);

    const handleFinished = useCallback(() => {
        setSuccess(true);

        setTimeout(() => {
            createSuccessNotification(
                translate("Successfully {{action}} the {{item}}", {
                    action: translate("added"),
                    item: translate("One-Time Password"),
                }),
            );

            props.setClosed();
            resetStates();
        }, 750);
    }, [createSuccessNotification, props, resetStates, translate]);

    const handleOnClose = () => {
        if (!props.open) {
            return;
        }

        handleClose();
    };

    useEffect(() => {
        if (!props.open || activeStep !== 0 || defaults !== null) {
            return;
        }

        (async () => {
            const opts = await getTOTPOptions();

            const decoded = {
                algorithm: toAlgorithmString(opts.algorithm),
                length: opts.length,
                period: opts.period,
            };

            setAvailable({
                algorithms: opts.algorithms.map((algorithm) => toAlgorithmString(algorithm)),
                lengths: opts.lengths,
                periods: opts.periods,
            });

            setDefaults(decoded);
            setSelected(decoded);
        })();
    }, [props.open, activeStep, defaults, selected]);

    const handleSetStepPrevious = useCallback(() => {
        if (activeStep === 0) {
            return;
        }

        setShowAdvanced(false);
        setActiveStep((prevState) => {
            return prevState - 1;
        });
    }, [activeStep]);

    const handleSetStepNext = useCallback(() => {
        if (activeStep === steps.length - 1) {
            return;
        }

        setShowAdvanced(false);
        setActiveStep((prevState) => {
            return prevState + 1;
        });
    }, [activeStep]);

    useEffect(() => {
        if (!props.open || activeStep !== 1) {
            return;
        }

        (async () => {
            setIsLoading(true);

            try {
                const secret = await getTOTPSecret(selected.algorithm, selected.length, selected.period);
                setSecretURL(secret.otpauth_url);
                setSecretValue(secret.base32_secret);
            } catch (err) {
                console.error(err);
                if ((err as Error).message.includes("Request failed with status code 403")) {
                    createErrorNotification(
                        translate("You must use the code from the same device and browser that initiated the process"),
                    );
                } else {
                    createErrorNotification(
                        translate("Failed to register device, the provided code is expired or has already been used"),
                    );
                }
                setHasErrored(true);
            }

            setIsLoading(false);
        })();
    }, [activeStep, createErrorNotification, selected, props.open, translate]);

    useEffect(() => {
        if (!props.open || activeStep !== 2 || dialState === State.InProgress || dialValue.length !== selected.length) {
            return;
        }

        (async () => {
            setDialState(State.InProgress);

            try {
                const registerValue = dialValue;
                setDialValue("");

                await completeTOTPRegister(registerValue);

                handleFinished();
            } catch (err) {
                console.error(err);
                setDialState(State.Failure);
            }
        })();
    }, [activeStep, dialState, dialValue, dialValue.length, handleFinished, props.open, selected.length]);

    const handleChangeAlgorithm = (ev: React.ChangeEvent<HTMLInputElement>, value: string) => {
        setSelected((prevState) => {
            return {
                ...prevState,
                algorithm: value,
            };
        });

        ev.preventDefault();
    };

    const handleChangeLength = (ev: React.ChangeEvent<HTMLInputElement>, value: string) => {
        setSelected((prevState) => {
            return {
                ...prevState,
                length: parseInt(value),
            };
        });

        ev.preventDefault();
    };

    const handleChangePeriod = (ev: React.ChangeEvent<HTMLInputElement>, value: string) => {
        setSelected((prevState) => {
            return {
                ...prevState,
                period: parseInt(value),
            };
        });

        ev.preventDefault();
    };

    const toggleAdvanced = () => {
        setShowAdvanced((prevState) => !prevState);
    };

    const advanced =
        defaults !== null &&
        (available.algorithms.length !== 1 || available.lengths.length !== 1 || available.periods.length !== 1);

    const disableAdvanced =
        defaults === null ||
        (available.algorithms.length <= 1 && available.lengths.length <= 1 && available.periods.length <= 1);

    const hideAlgorithms = advanced && available.algorithms.length <= 1;
    const hideLengths = advanced && available.lengths.length <= 1;
    const hidePeriods = advanced && available.periods.length <= 1;
    const qrcodeFuzzyStyle = isLoading || hasErrored ? classes.fuzzy : undefined;

    function renderStep(step: number) {
        switch (step) {
            case 0:
                return (
                    <Fragment>
                        {defaults === null ? (
                            <Grid size={{ xs: 12 }} my={3}>
                                <Typography>Loading...</Typography>
                            </Grid>
                        ) : (
                            <Grid container>
                                <Grid size={{ xs: 12 }} my={3}>
                                    <Typography>{translate("To begin select next")}</Typography>
                                </Grid>
                                <Grid size={{ xs: 12 }} hidden={disableAdvanced}>
                                    <FormControlLabel
                                        disabled={disableAdvanced}
                                        control={
                                            <Switch
                                                id={"one-time-password-advanced"}
                                                checked={showAdvanced}
                                                onChange={toggleAdvanced}
                                            />
                                        }
                                        label={translate("Advanced")}
                                    />
                                </Grid>
                                <Grid
                                    size={{ xs: 12 }}
                                    hidden={disableAdvanced || !showAdvanced}
                                    justifyContent={"center"}
                                    alignItems={"center"}
                                >
                                    <FormControl fullWidth>
                                        {hideAlgorithms ? null : (
                                            <Fragment>
                                                <FormLabel id={"lbl-adv-algorithms"}>
                                                    {translate("Algorithm")}
                                                </FormLabel>
                                                <RadioGroup
                                                    row
                                                    aria-labelledby={"lbl-adv-algorithms"}
                                                    value={selected.algorithm}
                                                    sx={{
                                                        justifyContent: "center",
                                                    }}
                                                    onChange={handleChangeAlgorithm}
                                                >
                                                    {available.algorithms.map((algorithm) => (
                                                        <FormControlLabel
                                                            key={algorithm}
                                                            value={algorithm}
                                                            control={
                                                                <Radio
                                                                    id={`one-time-password-algorithm-${algorithm}`}
                                                                />
                                                            }
                                                            label={algorithm}
                                                        />
                                                    ))}
                                                </RadioGroup>
                                            </Fragment>
                                        )}
                                        {hideLengths ? null : (
                                            <Fragment>
                                                <FormLabel id={"lbl-adv-lengths"}>{translate("Length")}</FormLabel>
                                                <RadioGroup
                                                    row
                                                    aria-labelledby={"lbl-adv-lengths"}
                                                    value={selected.length.toString()}
                                                    sx={{
                                                        justifyContent: "center",
                                                    }}
                                                    onChange={handleChangeLength}
                                                >
                                                    {available.lengths.map((length) => (
                                                        <FormControlLabel
                                                            key={length.toString()}
                                                            value={length.toString()}
                                                            control={
                                                                <Radio
                                                                    id={`one-time-password-length-${length.toString()}`}
                                                                />
                                                            }
                                                            label={length.toString()}
                                                        />
                                                    ))}
                                                </RadioGroup>
                                            </Fragment>
                                        )}
                                        {hidePeriods ? null : (
                                            <Fragment>
                                                <FormLabel id={"lbl-adv-periods"}>{translate("Seconds")}</FormLabel>
                                                <RadioGroup
                                                    row
                                                    aria-labelledby={"lbl-adv-periods"}
                                                    value={selected.period.toString()}
                                                    sx={{
                                                        justifyContent: "center",
                                                    }}
                                                    onChange={handleChangePeriod}
                                                >
                                                    {available.periods.map((period) => (
                                                        <FormControlLabel
                                                            key={period.toString()}
                                                            value={period.toString()}
                                                            control={
                                                                <Radio
                                                                    id={`one-time-password-period-${period.toString()}`}
                                                                />
                                                            }
                                                            label={period.toString()}
                                                        />
                                                    ))}
                                                </RadioGroup>
                                            </Fragment>
                                        )}
                                    </FormControl>
                                </Grid>
                            </Grid>
                        )}
                    </Fragment>
                );
            case 1:
                return (
                    <Fragment>
                        <Grid size={{ xs: 12 }} my={2}>
                            <FormControlLabel
                                control={
                                    <Switch
                                        id={"qr-toggle"}
                                        checked={showQRCode}
                                        onChange={() => {
                                            setShowQRCode((value) => !value);
                                        }}
                                    />
                                }
                                label={translate("QR Code")}
                            />
                        </Grid>
                        <Grid size={{ xs: 12 }} hidden={!showQRCode}>
                            <Box className={cx(qrcodeFuzzyStyle, classes.qrcodeContainer)}>
                                {secretURL !== null ? (
                                    <Link href={secretURL} underline="hover">
                                        <QRCodeSVG value={secretURL} className={classes.qrcode} size={200} />
                                        {!hasErrored && isLoading ? (
                                            <CircularProgress className={classes.loader} size={128} />
                                        ) : null}
                                        {hasErrored ? (
                                            <FontAwesomeIcon className={classes.failureIcon} icon={faTimesCircle} />
                                        ) : null}
                                    </Link>
                                ) : null}
                            </Box>
                        </Grid>
                        <Grid size={{ xs: 12 }} hidden={showQRCode}>
                            <Grid container spacing={2} justifyContent={"center"}>
                                <Grid size={{ xs: 4 }}>
                                    <CopyButton
                                        tooltip={translate("Click to Copy")}
                                        value={secretURL}
                                        childrenCopied={translate("Copied")}
                                        fullWidth
                                    >
                                        {translate("URI")}
                                    </CopyButton>
                                </Grid>
                                <Grid size={{ xs: 4 }}>
                                    <CopyButton
                                        tooltip={translate("Click to Copy")}
                                        value={secretValue}
                                        childrenCopied={translate("Copied")}
                                        fullWidth
                                    >
                                        {translate("Secret")}
                                    </CopyButton>
                                </Grid>
                                <Grid size={{ xs: 12 }}>
                                    <TextField
                                        id={"secret-url"}
                                        label={translate("Secret")}
                                        className={classes.secret}
                                        value={secretURL === null ? "" : secretURL}
                                        multiline={true}
                                        slotProps={{
                                            input: {
                                                readOnly: true,
                                            },
                                        }}
                                    />
                                </Grid>
                            </Grid>
                        </Grid>
                        <Grid size={{ xs: 12 }} sx={{ display: { xs: "none", md: "block" } }}>
                            <Box>
                                <Typography className={classes.googleAuthenticatorText}>
                                    {translate("Need Google Authenticator?")}
                                </Typography>
                                <AppStoreBadges
                                    iconSize={110}
                                    targetBlank
                                    className={classes.googleAuthenticatorBadges}
                                    googlePlayLink={GoogleAuthenticator.googlePlay}
                                    appleStoreLink={GoogleAuthenticator.appleStore}
                                />
                            </Box>
                        </Grid>
                    </Fragment>
                );
            case 2:
                return (
                    <Fragment>
                        <Grid size={{ xs: 12 }} paddingY={4}>
                            {success ? (
                                <Box className={classes.success}>
                                    <SuccessIcon />
                                </Box>
                            ) : (
                                <OTPDial
                                    passcode={dialValue}
                                    state={dialState}
                                    digits={selected.length}
                                    period={selected.period}
                                    onChange={setDialValue}
                                />
                            )}
                        </Grid>
                    </Fragment>
                );
        }
    }

    return (
        <Dialog open={props.open} onClose={handleOnClose} fullWidth>
            <DialogTitle>{translate("Register {{item}}", { item: translate("One-Time Password") })}</DialogTitle>
            <DialogContent>
                <DialogContentText sx={{ mb: 3 }}>
                    {translate("This dialog handles registration of a {{item}}", {
                        item: translate("One-Time Password"),
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
                    <Grid size={{ xs: 12 }}>
                        <Grid container spacing={1} justifyContent={"center"}>
                            {renderStep(activeStep)}
                        </Grid>
                    </Grid>
                </Grid>
            </DialogContent>
            <DialogActions>
                <Button
                    id={"dialog-previous"}
                    color={"primary"}
                    onClick={handleSetStepPrevious}
                    disabled={activeStep === 0}
                    data-1p-ignore
                >
                    {translate("Previous")}
                </Button>
                <Button id={"dialog-cancel"} color={"error"} onClick={handleClose} data-1p-ignore>
                    {translate("Cancel")}
                </Button>
                <Button
                    id={"dialog-next"}
                    color={"primary"}
                    onClick={handleSetStepNext}
                    disabled={activeStep === steps.length - 1}
                    data-1p-ignore
                >
                    {translate("Next")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    qrcode: {
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
        padding: theme.spacing(),
        backgroundColor: "white",
    },
    fuzzy: {
        filter: "blur(10px)",
    },
    secret: {
        marginTop: theme.spacing(1),
        marginBottom: theme.spacing(1),
        width: "256px",
    },
    googleAuthenticatorText: {
        fontSize: theme.typography.fontSize * 0.8,
    },
    googleAuthenticatorBadges: {},
    qrcodeContainer: {
        position: "relative",
        display: "inline-block",
    },
    loader: {
        position: "absolute",
        top: "calc(128px - 64px)",
        left: "calc(128px - 64px)",
        color: "rgba(255, 255, 255, 0.5)",
    },
    failureIcon: {
        position: "absolute",
        top: "calc(128px - 64px)",
        left: "calc(128px - 64px)",
        color: red[400],
        fontSize: "8rem",
    },
    success: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%",
    },
}));

export default OneTimePasswordRegisterDialog;
