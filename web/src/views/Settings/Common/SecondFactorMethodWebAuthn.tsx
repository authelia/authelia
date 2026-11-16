// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useReducer } from "react";

import axios from "axios";
import { useTranslation } from "react-i18next";

import WebAuthnTryIcon from "@components/WebAuthnTryIcon";
import { useAbortSignal } from "@hooks/Abort";
import { AssertionResult, AssertionResultFailureString, WebAuthnCeremony, WebAuthnTouchState } from "@models/WebAuthn";
import { getWebAuthnOptions, getWebAuthnResult, postWebAuthnResponse } from "@services/WebAuthn";

type ComponentState = {
    ceremony?: WebAuthnCeremony;
    status: WebAuthnTouchState;
    started: boolean;
};

type Action =
    | { type: "setCeremony"; ceremony?: WebAuthnCeremony }
    | { type: "setStarted"; started: boolean }
    | { type: "setStatus"; status: WebAuthnTouchState };

const initialState: ComponentState = {
    ceremony: undefined,
    started: false,
    status: WebAuthnTouchState.WaitTouch,
};

function reducer(state: ComponentState, action: Action): ComponentState {
    switch (action.type) {
        case "setCeremony":
            return { ...state, ceremony: action.ceremony };
        case "setStatus":
            return { ...state, status: action.status };
        case "setStarted":
            return { ...state, started: action.started };
        default:
            return state;
    }
}

export interface Props {
    onSecondFactorSuccess: () => void;
}

const SecondFactorMethodWebAuthn = function (props: Props) {
    const [state, dispatch] = useReducer(reducer, initialState);
    const getSignal = useAbortSignal();
    const { t: translate } = useTranslation();

    const { ceremony, started, status } = state;

    const handleRetry = () => {
        dispatch({ status: WebAuthnTouchState.WaitTouch, type: "setStatus" });
        dispatch({ started: false, type: "setStarted" });
    };

    const handleStart = useCallback(async () => {
        dispatch({ started: true, type: "setStarted" });
        dispatch({ ceremony: undefined, type: "setCeremony" });

        const signal = getSignal();

        try {
            const optionsStatus = await getWebAuthnOptions(signal);

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                dispatch({ status: WebAuthnTouchState.Failure, type: "setStatus" });
                console.error(new Error(translate("Unable to start sign in with your passkey or security key")));

                return;
            }

            dispatch({ ceremony: { timeout: optionsStatus.options.timeout }, type: "setCeremony" });

            const result = await getWebAuthnResult(optionsStatus.options);

            if (signal.aborted) return;

            if (result.result !== AssertionResult.Success) {
                dispatch({ status: WebAuthnTouchState.Failure, type: "setStatus" });

                console.error(new Error(translate(AssertionResultFailureString(result.result))));

                return;
            }

            if (result.response == null) {
                console.error(
                    new Error(translate("Your browser did not return a response from your passkey or security key")),
                );
                dispatch({ status: WebAuthnTouchState.Failure, type: "setStatus" });

                return;
            }

            dispatch({ status: WebAuthnTouchState.InProgress, type: "setStatus" });

            const response = await postWebAuthnResponse(
                result.response,
                undefined,
                undefined,
                undefined,
                undefined,
                undefined,
                signal,
            );

            if (response.data.status === "OK" && response.status === 200) {
                props.onSecondFactorSuccess();
                return;
            }

            console.error(new Error(translate("Your passkey or security key could not be verified")));
            dispatch({ status: WebAuthnTouchState.Failure, type: "setStatus" });
        } catch (err) {
            if (axios.isCancel(err)) return;
            console.error(err);
            dispatch({ status: WebAuthnTouchState.Failure, type: "setStatus" });
        }
    }, [getSignal, props, translate]);

    useEffect(() => {
        if (started) return;

        handleStart().catch(console.error);
    }, [handleStart, started]);

    return <WebAuthnTryIcon ceremony={ceremony} onRetryClick={handleRetry} webauthnTouchState={status} />;
};

export default SecondFactorMethodWebAuthn;
