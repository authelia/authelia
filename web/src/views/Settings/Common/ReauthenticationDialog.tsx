// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { lazy, useCallback, useEffect, useLayoutEffect, useMemo, useReducer, useRef } from "react";

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
import { Stepper } from "@components/UI/Stepper";
import { SecondFactorMethod } from "@models/Methods";
import { UserInfo } from "@models/UserInfo";
import { UserSessionElevation } from "@services/UserSessionElevation";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import ReauthenticationPasswordForm from "@views/Settings/Common/ReauthenticationPasswordForm";

const SecondFactorMethodMobilePush = lazy(() => import("@views/Settings/Common/SecondFactorMethodMobilePush"));
const SecondFactorMethodOneTimePassword = lazy(
    () => import("@views/Settings/Common/SecondFactorMethodOneTimePassword"),
);
const SecondFactorMethodWebAuthn = lazy(() => import("@views/Settings/Common/SecondFactorMethodWebAuthn"));

type Option = "password" | SecondFactorMethod;

type Props = {
    elevation?: UserSessionElevation;
    info?: UserInfo;
    opening: boolean;
    handleClosed: (_ok: boolean, _changed: boolean) => void;
    handleOpened: () => void;
};

type State = {
    open: boolean;
    closing: boolean;
    activeStep: number;
    option: Option | undefined;
};

type Action =
    | { type: "reset" }
    | { type: "setActiveStep"; payload: number }
    | { type: "setClosing"; payload: boolean }
    | { type: "setOpen"; payload: boolean }
    | { type: "setOption"; payload: Option | undefined };

const initialState: State = {
    activeStep: 0,
    closing: false,
    open: false,
    option: undefined,
};

function reducer(state: State, action: Action): State {
    switch (action.type) {
        case "reset":
            return { ...initialState };
        case "setOpen":
            return { ...state, open: action.payload };
        case "setClosing":
            return { ...state, closing: action.payload };
        case "setActiveStep":
            return { ...state, activeStep: action.payload };
        case "setOption":
            return { ...state, option: action.payload };
        default:
            return state;
    }
}

export function getReauthenticationOptions(elevation: UserSessionElevation, info: UserInfo): Option[] {
    const options: Option[] = [];

    if (elevation.reauthentication_methods.includes("password")) {
        options.push("password");
    }

    if (elevation.reauthentication_methods.includes("second_factor")) {
        if (info.has_totp) options.push(SecondFactorMethod.TOTP);
        if (info.has_webauthn && browserSupportsWebAuthn()) options.push(SecondFactorMethod.WebAuthn);
        if (info.has_duo) options.push(SecondFactorMethod.MobilePush);
    }

    return options;
}

const ReauthenticationDialog = function (props: Props) {
    const { elevation, handleClosed, handleOpened, info, opening } = props;
    const { t: translate } = useTranslation(["settings", "portal"]);

    const [state, dispatch] = useReducer(reducer, initialState);
    const { activeStep, closing, open, option } = state;

    const timeoutSuccessRef = useRef<null | ReturnType<typeof setTimeout>>(null);

    useEffect(() => {
        return () => {
            if (timeoutSuccessRef.current !== null) {
                clearTimeout(timeoutSuccessRef.current);
                timeoutSuccessRef.current = null;
            }
        };
    }, []);

    const options = useMemo(
        () => (elevation && info ? getReauthenticationOptions(elevation, info) : []),
        [elevation, info],
    );

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

    const handleSelect = useCallback(
        (selected: Option) => {
            if (closing) return;

            dispatch({ payload: selected, type: "setOption" });
            dispatch({ payload: 1, type: "setActiveStep" });
        },
        [closing],
    );

    const handleSuccess = useCallback(() => {
        dispatch({ payload: true, type: "setClosing" });
        dispatch({ payload: 2, type: "setActiveStep" });

        timeoutSuccessRef.current = setTimeout(() => {
            timeoutSuccessRef.current = null;

            handleClose(true, true);
        }, 1500);
    }, [handleClose]);

    const handleCancelled = () => {
        // A success is already scheduled to close the dialog, so a cancellation must not report a conflicting result.
        if (closing) return;

        handleClose(false, false);
    };

    useLayoutEffect(() => {
        if (closing || !opening || !elevation) return;

        if (!elevation.require_reauthentication) {
            resetState();
            handleClosed(true, false);
            return;
        }

        if (!open) {
            handleOpened();
            dispatch({ payload: true, type: "setOpen" });
        }

        if (options.length === 1 && option === undefined) {
            dispatch({ payload: options[0], type: "setOption" });
            dispatch({ payload: 1, type: "setActiveStep" });
        }
    }, [closing, elevation, resetState, handleClosed, open, opening, option, options, handleOpened]);

    const renderMethod = () => {
        switch (option) {
            case "password":
                return <ReauthenticationPasswordForm onAuthenticationSuccess={handleSuccess} />;
            case SecondFactorMethod.TOTP:
                return <SecondFactorMethodOneTimePassword onSecondFactorSuccess={handleSuccess} />;
            case SecondFactorMethod.WebAuthn:
                return <SecondFactorMethodWebAuthn onSecondFactorSuccess={handleSuccess} />;
            case SecondFactorMethod.MobilePush:
                return <SecondFactorMethodMobilePush onSecondFactorSuccess={handleSuccess} />;
            default:
                return null;
        }
    };

    const getLabel = (value: Option) => {
        switch (value) {
            case "password":
                return translate("Password");
            case SecondFactorMethod.TOTP:
                return translate("One-Time Password");
            case SecondFactorMethod.WebAuthn:
                return translate("WebAuthn");
            case SecondFactorMethod.MobilePush:
                return translate("Mobile Push");
            default:
                return value;
        }
    };

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

        if (options.length === 0) {
            return (
                <p className="my-16 text-center">{translate("There are no methods available to authenticate again")}</p>
            );
        }

        if (activeStep === 0) {
            return (
                <div className="flex flex-col items-center justify-center gap-4 my-16">
                    {options.map((value) => (
                        <Button key={value} variant={"outline"} onClick={() => handleSelect(value)}>
                            {getLabel(value)}
                        </Button>
                    ))}
                </div>
            );
        }

        return <div className="flex flex-col items-center justify-center my-16">{renderMethod()}</div>;
    };

    return (
        <Dialog
            open={open}
            onOpenChange={(isOpen) => {
                if (!isOpen) handleCancelled();
            }}
        >
            <DialogContent id={"dialog-reauthenticate"} showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{translate("Identity Verification")}</DialogTitle>
                    <DialogDescription>
                        {translate(
                            "In order to perform this action, policy enforcement requires that you authenticate again",
                        )}
                    </DialogDescription>
                </DialogHeader>
                <Stepper
                    activeStep={activeStep}
                    steps={[translate("Select a Method"), translate("Authenticate"), translate("Completed")]}
                />
                {renderContent()}
                <DialogFooter>
                    <Button variant={"outline"} color={"destructive"} disabled={closing} onClick={handleCancelled}>
                        {translate("Cancel")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default ReauthenticationDialog;
