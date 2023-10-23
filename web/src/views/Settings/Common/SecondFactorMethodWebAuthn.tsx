import React, { useCallback, useEffect, useState } from "react";

import WebAuthnTryIcon from "@components/WebAuthnTryIcon";
import { useIsMountedRef } from "@hooks/Mounted";
import { AssertionResult, AssertionResultFailureString, WebAuthnTouchState } from "@models/WebAuthn";
import { getAuthenticationOptions, getAuthenticationResult, postAuthenticationResponse } from "@services/WebAuthn";

export interface Props {
    closing: boolean;
    onSecondFactorSuccess: () => void;
}

const SecondFactorMethodWebAuthn = function (props: Props) {
    const [state, setState] = useState(WebAuthnTouchState.WaitTouch);
    const [started, setStarted] = useState(false);
    const mounted = useIsMountedRef();

    const handleRetry = () => {
        setState(WebAuthnTouchState.WaitTouch);
    };

    const handleStart = useCallback(async () => {
        setStarted(true);

        try {
            const optionsStatus = await getAuthenticationOptions();

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                setState(WebAuthnTouchState.Failure);
                console.error(new Error("Failed to initiate security key sign in process"));

                return;
            }

            const result = await getAuthenticationResult(optionsStatus.options);

            if (result.result !== AssertionResult.Success) {
                if (!mounted.current) return;

                setState(WebAuthnTouchState.Failure);

                console.error(new Error(AssertionResultFailureString(result.result)));

                return;
            }

            if (result.response == null) {
                console.error(new Error("The browser did not respond with the expected attestation data."));
                setState(WebAuthnTouchState.Failure);

                return;
            }

            if (!mounted.current) return;

            setState(WebAuthnTouchState.InProgress);

            const response = await postAuthenticationResponse(result.response);

            if (response.data.status === "OK" && response.status === 200) {
                props.onSecondFactorSuccess();
                return;
            }

            if (!mounted.current) return;

            console.error(new Error("The server rejected the security key."));
            setState(WebAuthnTouchState.Failure);
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            // onSignInErrorCallback(new Error("Failed to initiate security key sign in process"));
            setState(WebAuthnTouchState.Failure);
        }
    }, [mounted, props]);

    useEffect(() => {
        if (started) return;

        handleStart().catch(console.error);
    }, [handleStart, started]);

    return <WebAuthnTryIcon onRetryClick={handleRetry} webauthnTouchState={state} />;
};

export default SecondFactorMethodWebAuthn;
