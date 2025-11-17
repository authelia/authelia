import { Fragment, lazy, useCallback, useLayoutEffect, useReducer } from "react";

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

type State = {
    open: boolean;
    loading: boolean;
    closing: boolean;
    activeStep: number;
    method: SecondFactorMethod | undefined;
};

type Action =
    | { type: "reset" }
    | { type: "setActiveStep"; payload: number }
    | { type: "setClosing"; payload: boolean }
    | { type: "setLoading"; payload: boolean }
    | { type: "setMethod"; payload: SecondFactorMethod | undefined }
    | { type: "setOpen"; payload: boolean };

const initialState: State = {
    activeStep: 0,
    closing: false,
    loading: false,
    method: undefined,
    open: false,
};

function reducer(state: State, action: Action): State {
    switch (action.type) {
        case "reset":
            return { ...initialState };
        case "setOpen":
            return { ...state, open: action.payload };
        case "setLoading":
            return { ...state, loading: action.payload };
        case "setClosing":
            return { ...state, closing: action.payload };
        case "setActiveStep":
            return { ...state, activeStep: action.payload };
        case "setMethod":
            return { ...state, method: action.payload };
        default:
            return state;
    }
}

const SecondFactorDialog = function (props: Props) {
    const { elevation, handleClosed, handleOpened, info, opening } = props;
    const { t: translate } = useTranslation(["settings", "portal"]);
    const { classes } = useStyles();

    const [state, dispatch] = useReducer(reducer, initialState);
    const { activeStep, closing, loading, method, open } = state;

    const resetState = useCallback(() => {
        dispatch({ type: "reset" });
    }, []);

    const handleClose = useCallback(
        (ok: boolean, changed: boolean) => {
            resetState();
            handleClosed(ok, changed);
        },
        [resetState, handleClosed],
    );

    const handleCancelled = () => {
        handleClose(false, false);
    };

    const handleOneTimeCode = () => {
        handleClose(true, false);
    };

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

        dispatch({ payload: method, type: "setMethod" });
        dispatch({ payload: 1, type: "setActiveStep" });
    };

    const handleSuccess = useCallback(() => {
        dispatch({ payload: true, type: "setClosing" });
        dispatch({ payload: 2, type: "setActiveStep" });

        setTimeout(() => {
            handleClose(true, true);
        }, 1500);
    }, [handleClose]);

    useLayoutEffect(() => {
        if (closing || !opening || !elevation) return;

        const shouldSkip =
            (elevation.skip_second_factor || !elevation.require_second_factor) && !elevation.can_skip_second_factor;
        if (shouldSkip) {
            resetState();
            handleClosed(true, false);
            return;
        }

        if (!open) {
            handleOpened();
            dispatch({ payload: true, type: "setOpen" });
        }

        if (!elevation.factor_knowledge) {
            dispatch({ payload: 1, type: "setActiveStep" });
        }
    }, [closing, resetState, handleClosed, open, elevation, opening, handleOpened]);

    const getAuthComponent = useCallback(() => {
        if (!elevation?.factor_knowledge) {
            return <PasswordForm onAuthenticationSuccess={handleSuccess} />;
        }

        switch (method) {
            case SecondFactorMethod.WebAuthn:
                return <SecondFactorMethodWebAuthn onSecondFactorSuccess={handleSuccess} />;
            case SecondFactorMethod.TOTP:
                return <SecondFactorMethodOneTimePassword onSecondFactorSuccess={handleSuccess} />;
            case SecondFactorMethod.MobilePush:
                return <SecondFactorMethodMobilePush onSecondFactorSuccess={handleSuccess} />;
            default:
                return null;
        }
    }, [elevation, method, handleSuccess]);

    const renderContent = () => {
        if (activeStep === 2) {
            return (
                <Box
                    className={classes.success}
                    sx={{
                        display: "flex",
                        flexDirection: "column",
                        m: "auto",
                        padding: "5.0rem",
                        width: "fit-content",
                    }}
                >
                    <SuccessIcon />
                </Box>
            );
        }

        if (!elevation || !info) {
            return <LoadingPage />;
        }

        if (activeStep === 0) {
            return (
                <Stack alignContent={"center"} justifyContent={"center"} alignItems={"center"} spacing={2} my={8}>
                    {elevation.can_skip_second_factor ? (
                        <Fragment>
                            <Button variant={"outlined"} onClick={handleOneTimeCode}>
                                {translate("Email One-Time Code")}
                            </Button>
                            <Divider />
                            <Typography variant={"h5"}>{translate("or", { ns: "portal" })}</Typography>
                            <Divider />
                        </Fragment>
                    ) : null}
                    {info.has_totp ? (
                        <Button variant={"outlined"} onClick={handleClickOneTimePassword}>
                            {translate("One-Time Password")}
                        </Button>
                    ) : null}
                    {info.has_webauthn && browserSupportsWebAuthn() ? (
                        <Button variant={"outlined"} onClick={handleClickWebAuthn}>
                            {translate("WebAuthn")}
                        </Button>
                    ) : null}
                    {info.has_duo ? (
                        <Button variant={"outlined"} onClick={handleClickMobilePush}>
                            {translate("Mobile Push")}
                        </Button>
                    ) : null}
                </Stack>
            );
        }

        if (activeStep === 1) {
            return (
                <Stack alignContent={"center"} justifyContent={"center"} alignItems={"center"} my={8}>
                    {getAuthComponent()}
                </Stack>
            );
        }

        return <LoadingPage />;
    };

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
                {renderContent()}
            </DialogContent>
            <DialogActions>
                <Button variant={"outlined"} color={"error"} disabled={loading} onClick={handleCancelled}>
                    {translate("Cancel")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    success: {
        display: "flex",
        flex: "0 0 100%",
        flexDirection: "column",
        m: "auto",
        marginBottom: theme.spacing(2),
        marginY: "2.5rem",
        width: "fit-content",
    },
}));

export default SecondFactorDialog;
