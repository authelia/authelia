import { useCallback, useEffect, useReducer } from "react";

import { useTranslation } from "react-i18next";

import WebAuthnTryIcon from "@components/WebAuthnTryIcon";
import { useIsMountedRef } from "@hooks/Mounted";
import { AssertionResult, AssertionResultFailureString, WebAuthnTouchState } from "@models/WebAuthn";
import { getWebAuthnOptions, getWebAuthnResult, postWebAuthnResponse } from "@services/WebAuthn";

type ComponentState = {
    status: WebAuthnTouchState;
    started: boolean;
};

type Action = { type: "setStarted"; started: boolean } | { type: "setStatus"; status: WebAuthnTouchState };

const initialState: ComponentState = {
    started: false,
    status: WebAuthnTouchState.WaitTouch,
};

function reducer(state: ComponentState, action: Action): ComponentState {
    switch (action.type) {
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
    const mounted = useIsMountedRef();
    const { t: translate } = useTranslation();

    const { started, status } = state;

    const handleRetry = () => {
        dispatch({ status: WebAuthnTouchState.WaitTouch, type: "setStatus" });
        dispatch({ started: false, type: "setStarted" });
    };

    const handleStart = useCallback(async () => {
        dispatch({ started: true, type: "setStarted" });

        try {
            const optionsStatus = await getWebAuthnOptions();

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                dispatch({ status: WebAuthnTouchState.Failure, type: "setStatus" });
                console.error(new Error(translate("Failed to initiate security key sign in process")));

                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (result.result !== AssertionResult.Success) {
                if (!mounted.current) return;

                dispatch({ status: WebAuthnTouchState.Failure, type: "setStatus" });

                console.error(new Error(translate(AssertionResultFailureString(result.result))));

                return;
            }

            if (result.response == null) {
                console.error(new Error(translate("The browser did not respond with the expected attestation data")));
                dispatch({ status: WebAuthnTouchState.Failure, type: "setStatus" });

                return;
            }

            if (!mounted.current) return;

            dispatch({ status: WebAuthnTouchState.InProgress, type: "setStatus" });

            const response = await postWebAuthnResponse(result.response);

            if (response.data.status === "OK" && response.status === 200) {
                props.onSecondFactorSuccess();
                return;
            }

            if (!mounted.current) return;

            console.error(new Error(translate("The server rejected the security key")));
            dispatch({ status: WebAuthnTouchState.Failure, type: "setStatus" });
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            dispatch({ status: WebAuthnTouchState.Failure, type: "setStatus" });
        }
    }, [mounted, props, translate]);

    useEffect(() => {
        if (started) return;

        handleStart().catch(console.error);
    }, [handleStart, started]);

    return <WebAuthnTryIcon onRetryClick={handleRetry} webauthnTouchState={status} />;
};

export default SecondFactorMethodWebAuthn;
