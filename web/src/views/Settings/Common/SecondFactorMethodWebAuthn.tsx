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

type Action = { type: "setStatus"; status: WebAuthnTouchState } | { type: "setStarted"; started: boolean };

const initialState: ComponentState = {
    status: WebAuthnTouchState.WaitTouch,
    started: false,
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

    const { status, started } = state;

    const handleRetry = () => {
        dispatch({ type: "setStatus", status: WebAuthnTouchState.WaitTouch });
    };

    const handleStart = useCallback(async () => {
        dispatch({ type: "setStarted", started: true });

        try {
            const optionsStatus = await getWebAuthnOptions();

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                dispatch({ type: "setStatus", status: WebAuthnTouchState.Failure });
                console.error(new Error(translate("Failed to initiate security key sign in process")));

                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (result.result !== AssertionResult.Success) {
                if (!mounted.current) return;

                dispatch({ type: "setStatus", status: WebAuthnTouchState.Failure });

                console.error(new Error(translate(AssertionResultFailureString(result.result))));

                return;
            }

            if (result.response == null) {
                console.error(new Error(translate("The browser did not respond with the expected attestation data")));
                dispatch({ type: "setStatus", status: WebAuthnTouchState.Failure });

                return;
            }

            if (!mounted.current) return;

            dispatch({ type: "setStatus", status: WebAuthnTouchState.InProgress });

            const response = await postWebAuthnResponse(result.response);

            if (response.data.status === "OK" && response.status === 200) {
                props.onSecondFactorSuccess();
                return;
            }

            if (!mounted.current) return;

            console.error(new Error(translate("The server rejected the security key")));
            dispatch({ type: "setStatus", status: WebAuthnTouchState.Failure });
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            dispatch({ type: "setStatus", status: WebAuthnTouchState.Failure });
        }
    }, [mounted, props, translate]);

    useEffect(() => {
        if (started) return;

        handleStart().catch(console.error);
    }, [handleStart, started]);

    return <WebAuthnTryIcon onRetryClick={handleRetry} webauthnTouchState={status} />;
};

export default SecondFactorMethodWebAuthn;
