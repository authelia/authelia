import React, { Fragment, lazy, useCallback, useEffect, useState } from "react";

import {
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Divider,
    Stack,
    Step,
    StepLabel,
    Stepper,
    Theme,
    Typography,
} from "@mui/material";
import { browserSupportsWebAuthn } from "@simplewebauthn/browser";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import SuccessIcon from "@components/SuccessIcon";
import { SecondFactorMethod } from "@models/Methods";
import { UserInfo } from "@models/UserInfo";
import { UserSessionElevation } from "@services/UserSessionElevation";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import PasswordForm from "@views/LoginPortal/SecondFactor/PasswordForm";

const SecondFactorMethodMobilePush = lazy(() => import("@views/Settings/Common/SecondFactorMethodMobilePush"));
const SecondFactorMethodOneTimePassword = lazy(
    () => import("@views/Settings/Common/SecondFactorMethodOneTimePassword"),
);
const SecondFactorMethodWebAuthn = lazy(() => import("@views/Settings/Common/SecondFactorMethodWebAuthn"));

type Props = {
    elevation?: UserSessionElevation;
    info?: UserInfo;
    opening: boolean;
    handleClosed: (ok: boolean, changed: boolean) => void;
    handleOpened: () => void;
};

const SecondFactorDialog = function (props: Props) {
    const { t: translate } = useTranslation(["settings", "portal"]);
    const { classes } = useStyles();

    const [open, setOpen] = useState(false);
    const [loading, setLoading] = useState(false);
    const [closing, setClosing] = useState(false);
    const [activeStep, setActiveStep] = useState(0);
    const [method, setMethod] = useState<SecondFactorMethod>();

    const handleClose = useCallback(
        (ok: boolean, changed: boolean) => {
            setOpen(false);

            setActiveStep(0);

            setLoading(false);
            setClosing(false);
            setMethod(undefined);
            props.handleClosed(ok, changed);
        },
        [props],
    );

    const handleCancelled = () => {
        handleClose(false, false);
    };

    const handleOneTimeCode = () => {
        handleClose(true, false);
    };

    const handleLoad = useCallback(async () => {
        if (closing || !props.elevation) return;

        if (
            (props.elevation.skip_second_factor || !props.elevation.require_second_factor) &&
            !props.elevation.can_skip_second_factor
        ) {
            handleClose(true, false);

            return;
        }

        if (!open) {
            props.handleOpened();
            setOpen(true);
        }

        if (!props.elevation.factor_knowledge) {
            setActiveStep(1);
        }
    }, [closing, handleClose, open, props]);

    const handleClickOneTimePassword = () => {
        handleClick(SecondFactorMethod.TOTP);
    };

    const handleClickWebAuthn = () => {
        handleClick(SecondFactorMethod.WebAuthn);
    };

    const handleClickMobilePush = () => {
        handleClick(SecondFactorMethod.MobilePush);
    };

    const handleClick = (method: SecondFactorMethod) => {
        if (closing) return;

        setMethod(method);

        setActiveStep(1);
    };

    const handleSuccess = () => {
        setClosing(true);

        setActiveStep(2);

        setTimeout(() => {
            handleClose(true, true);
        }, 1500);
    };

    useEffect(() => {
        if (closing || !props.opening || !props.elevation) return;

        handleLoad().catch(console.error);
    }, [closing, handleLoad, props, props.opening]);

    return (
        <Dialog id={"dialog-verify-second-factor"} open={open} onClose={handleCancelled}>
            <DialogTitle>{translate("Identity Verification")}</DialogTitle>
            <DialogContent>
                <DialogContentText gutterBottom>
                    {translate(
                        "In order to perform this action, policy enforcement requires that two-factor authentication is performed",
                    )}
                </DialogContentText>
                <Stepper activeStep={activeStep}>
                    <Step key={"step-1"}>
                        <StepLabel>{translate("Select a Method")}</StepLabel>
                    </Step>
                    <Step key={"step-2"}>
                        <StepLabel>{translate("Authenticate")}</StepLabel>
                    </Step>
                    <Step key={"step-3"}>
                        <StepLabel>{translate("Completed")}</StepLabel>
                    </Step>
                </Stepper>
                {!props.elevation || !props.info ? (
                    activeStep === 2 ? (
                        <Box
                            className={classes.success}
                            sx={{
                                display: "flex",
                                flexDirection: "column",
                                m: "auto",
                                width: "fit-content",
                                padding: "5.0rem",
                            }}
                        >
                            <SuccessIcon />
                        </Box>
                    ) : (
                        <LoadingPage />
                    )
                ) : activeStep === 0 ? (
                    <Stack alignContent={"center"} justifyContent={"center"} alignItems={"center"} spacing={2} my={8}>
                        {props.elevation.can_skip_second_factor ? (
                            <Fragment>
                                <Button variant={"outlined"} onClick={handleOneTimeCode} data-1p-ignore>
                                    {translate("Email One-Time Code")}
                                </Button>
                                <Divider />
                                <Typography variant={"h5"}>{translate("or", { ns: "portal" })}</Typography>
                                <Divider />
                            </Fragment>
                        ) : null}
                        {props.info.has_totp ? (
                            <Button variant={"outlined"} onClick={handleClickOneTimePassword} data-1p-ignore>
                                {translate("One-Time Password")}
                            </Button>
                        ) : null}
                        {props.info.has_webauthn && browserSupportsWebAuthn() ? (
                            <Button variant={"outlined"} onClick={handleClickWebAuthn} data-1p-ignore>
                                {translate("WebAuthn")}
                            </Button>
                        ) : null}
                        {props.info.has_duo ? (
                            <Button variant={"outlined"} onClick={handleClickMobilePush} data-1p-ignore>
                                {translate("Mobile Push")}
                            </Button>
                        ) : null}
                    </Stack>
                ) : activeStep === 1 ? (
                    <Stack alignContent={"center"} justifyContent={"center"} alignItems={"center"} my={8}>
                        {!props.elevation.factor_knowledge ? (
                            <PasswordForm onAuthenticationSuccess={handleSuccess} />
                        ) : method === SecondFactorMethod.WebAuthn ? (
                            <SecondFactorMethodWebAuthn onSecondFactorSuccess={handleSuccess} closing={closing} />
                        ) : method === SecondFactorMethod.TOTP ? (
                            <SecondFactorMethodOneTimePassword
                                onSecondFactorSuccess={handleSuccess}
                                closing={closing}
                            />
                        ) : method === SecondFactorMethod.MobilePush ? (
                            <SecondFactorMethodMobilePush onSecondFactorSuccess={handleSuccess} closing={closing} />
                        ) : null}
                    </Stack>
                ) : (
                    <Box
                        className={classes.success}
                        sx={{
                            display: "flex",
                            flexDirection: "column",
                            m: "auto",
                            width: "fit-content",
                            padding: "5.0rem",
                        }}
                    >
                        <SuccessIcon />
                    </Box>
                )}
            </DialogContent>
            <DialogActions>
                <Button
                    variant={"outlined"}
                    color={"error"}
                    disabled={loading}
                    onClick={handleCancelled}
                    data-1p-ignore
                >
                    {translate("Cancel")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    success: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%",
        display: "flex",
        flexDirection: "column",
        m: "auto",
        width: "fit-content",
        marginY: "2.5rem",
    },
}));

export default SecondFactorDialog;
