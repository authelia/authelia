import { Fragment, lazy, useCallback, useLayoutEffect, useReducer } from "react";

import { browserSupportsWebAuthn } from "@simplewebauthn/browser";
import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";
import { Button } from "@components/UI/Button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@components/UI/Dialog";
import { Separator } from "@components/UI/Separator";
import { Step, StepLabel, Stepper } from "@components/UI/Stepper";
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
    handleClosed: (_ok: boolean, _changed: boolean) => void;
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
                <div className="flex flex-col m-auto p-20 w-fit">
                    <SuccessIcon />
                </div>
            );
        }

        if (!elevation || !info) {
            return <LoadingPage />;
        }

        if (activeStep === 0) {
            return (
                <div className="flex flex-col items-center justify-center gap-4 my-16">
                    {elevation.can_skip_second_factor ? (
                        <Fragment>
                            <Button variant={"outline"} onClick={handleOneTimeCode}>
                                {translate("Email One-Time Code")}
                            </Button>
                            <Separator />
                            <h5 className="text-xl font-semibold">{translate("or", { ns: "portal" })}</h5>
                            <Separator />
                        </Fragment>
                    ) : null}
                    {info.has_totp ? (
                        <Button variant={"outline"} onClick={handleClickOneTimePassword}>
                            {translate("One-Time Password")}
                        </Button>
                    ) : null}
                    {info.has_webauthn && browserSupportsWebAuthn() ? (
                        <Button variant={"outline"} onClick={handleClickWebAuthn}>
                            {translate("WebAuthn")}
                        </Button>
                    ) : null}
                    {info.has_duo ? (
                        <Button variant={"outline"} onClick={handleClickMobilePush}>
                            {translate("Mobile Push")}
                        </Button>
                    ) : null}
                </div>
            );
        }

        if (activeStep === 1) {
            return <div className="flex flex-col items-center justify-center my-16">{getAuthComponent()}</div>;
        }

        return <LoadingPage />;
    };

    return (
        <Dialog
            open={open}
            onOpenChange={(isOpen) => {
                if (!isOpen) handleCancelled();
            }}
        >
            <DialogContent id={"dialog-verify-second-factor"} showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{translate("Identity Verification")}</DialogTitle>
                    <DialogDescription>
                        {translate(
                            "In order to perform this action, policy enforcement requires that two-factor authentication is performed",
                        )}
                    </DialogDescription>
                </DialogHeader>
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
                <DialogFooter>
                    <Button variant={"outline"} color={"destructive"} disabled={loading} onClick={handleCancelled}>
                        {translate("Cancel")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default SecondFactorDialog;
